package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// struct.go contains data structures for modeling a golang struct: Struct, Property, and Tag

// structTemplate outlines how a Struct model is rendered into a Golang type definition
var structTemplate = template.Must(template.New("").Funcs(DefaultFuncMap).Parse(`
// {{.Name}} ... {{ if .DocumentationURL }}
// {{.DocumentationURL}} {{end}}{{if .Documentation}}
//
{{.DocComment}}{{end}}
type {{.Name}} struct {
	{{- range .Properties }}
	// {{.StructFieldName}} ... {{ if .Documentation }}
	//
	{{.DocComment}}{{- end}}
	{{.StructFieldName}} {{.Type}} {{.Tag}}
	{{- end}}
}`))

//
var embeddedStructTemplate = template.Must(template.New("").Funcs(DefaultFuncMap).Parse(`
// {{.Name}} ... {{ if .DocumentationURL}}
// {{.DocumentationURL}}
//{{end}}{{if .Documentation}}
//
// {{.DocComment}}{{end}}
type {{.Name}} struct {
	{{- range .Properties }}
	// {{.StructFieldName}} ... {{ if .Documentation }}
	//
	{{.DocComment}}{{- end}}
	{{.StructFieldName}} 
	{{- end}}
}`))

// the salesforce REST API returns child relationships as nested json of type QueryResponse
// access becomes similar to Lead.Contacts.Records[0] where the Contacts is the QueryResponse
//
// ideally this JSON might be flattened or manipulated for easier access but for now I'm choosing
// to mirror the model of the API response
var childTypeTemplate = template.Must(template.New("").
	Funcs(DefaultFuncMap).
	Parse("struct {\n\tDone bool `json:\"done\"`\n\tCount int `json:\"count\"`\n\tTotalSize int `json:\"totalSize\"`\n\tRecords []*{{.}} `json:\"records\"`\n}"),
)

// Struct contains the components to generate a golang struct definition
type Struct struct {
	// DocumentationURL ...
	DocumentationURL string
	// ParentName ...
	ParentName string
	// Name ...
	Name string
	// Documentation ...
	Documentation string
	// Properties respresent a struct field
	Properties Properties
	// Dependencies are salesforce objects required by child or parent relationships of this struct
	Dependencies []string
	// IsPlolymorphicModel represents a model with two or more embeded structs representing a polymorphic relationship
	IsPolymorphicModel bool
}

func (s *Struct) String() (string, error) {
	buffer := &bytes.Buffer{}

	if s.IsPolymorphicModel {
		if err := embeddedStructTemplate.Execute(buffer, s); err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	if err := structTemplate.Execute(buffer, s); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// Valid ...
func (s *Struct) Valid() bool {
	if s.Name == "" {
		return false
	}

	for _, d := range s.Dependencies {
		if d == "" {
			return false
		}
	}

	for _, p := range s.Properties {
		if p.Name == "" {
			return false
		}

		if p.Type == "" && !p.IsEmbedded {
			return false
		}
	}

	return true
}

// RemoveRelations returns a new copy of Struct with properties representing
// parent or child relations removed
func (s *Struct) RemoveRelations() *Struct {
	var nonRelations []*Property

	for _, property := range s.Properties {
		if property.IsRelation() {
			continue
		}

		nonRelations = append(nonRelations, property)
	}

	return &Struct{
		ParentName:       s.ParentName,
		Name:             s.Name,
		DocumentationURL: s.DocumentationURL,
		Documentation:    s.Documentation,
		Properties:       nonRelations,
		Dependencies:     nil,
	}
}

func (s *Struct) DocComment() string {
	ss := strings.Split(strings.ReplaceAll(s.Documentation, "\r\n", "\n"), "\n")
	out := ""
	for i, s := range ss {
		if i > 0 {
			out = fmt.Sprintf("%s\n// %s", out, s)
		} else {
			out = fmt.Sprintf("%s// %s", out, s)
		}
	}
	return out
}

// Property represents a Field of a struct
type Property struct {
	// ParentName is a salesforce specific annotation used to indicate
	// the parent SObject in a parent:child relationship
	ParentName string
	// Name ...
	Name string
	// Documentation ...
	Documentation string
	// Type ...
	Type string
	// Tag ...
	Tag Tag
	// IsEmbedded indicated an embedded property such as
	//
	// type Foo struct {
	//     Bar
	// }
	//
	// This is used to provide response types for polymorphic foreign keys
	IsEmbedded bool
	// IsNillable ...
	IsNillable bool
}

// NewProperty ...
func NewProperty(parent string, name string, dataType string, documentation string, tagName string, tagValues ...string) (Property, error) {
	prop := Property{}

	var err error
	prop.Type, err = convertType(dataType)
	if err != nil {
		return prop, err
	}

	if parent != "" {
		prop.ParentName, err = prepareStructName(parent, singular)
		if err != nil {
			return prop, err
		}
	}

	prop.Name, err = prepareStructName(name, neither)
	if err != nil {
		return prop, err
	}

	prop.Documentation = documentation //enforceLineLimit(documentation, 90)
	prop.Tag = Tag{tagName: tagValues}
	return prop, nil
}

// NewParentProperty ...
func NewParentProperty(parentName string, relationshipName string) Property {
	return Property{
		ParentName: parentName,
		Name:       parentName,
		Type:       "*" + parentName,
		Tag:        jsonTag(relationshipName),
	}
}

// NewChildProperty ...
func NewChildProperty(parentName string, relationshipName string) (Property, error) {
	var buff bytes.Buffer
	if err := childTypeTemplate.Execute(&buff, parentName); err != nil {
		return Property{}, fmt.Errorf("executing relationTypeTemplate: %w", err)
	}

	return Property{
		ParentName: parentName,
		Name:       relationshipName,
		Type:       buff.String(),
		Tag:        jsonTag(relationshipName),
	}, nil
}

// IsRelation returns true if this property represents a parent or child relationship
func (p *Property) IsRelation() bool {
	return p.ParentName != "" || p.IsEmbedded
}

func (p *Property) StructFieldName() string {
	const f = "%s%s"
	return fmt.Sprintf(f, strings.ToUpper(string(p.Name[0])), p.Name[1:])
}

func (p *Property) DocComment() string {
	ss := strings.Split(strings.ReplaceAll(p.Documentation, "\r\n", "\n"), "\n")
	out := ""
	for i, s := range ss {
		if i > 0 {
			out = fmt.Sprintf("%s\n// %s", out, s)
		} else {
			out = fmt.Sprintf("%s// %s", out, s)
		}
	}
	return out
}

// Tag represents a set of struct property tags
// i.e. `json:"foo,omitempty" bson:"bar"`
type Tag map[string][]string

func jsonTag(values ...string) Tag {
	return Tag{
		"json": values,
	}
}

// Add extends the key-value pairing for a golang JSON struct tag
func (t Tag) Add(key string, value string) {
	if values, exists := t[key]; exists {
		t[key] = append(values, value)
	} else {
		t[key] = []string{value}
	}
}

// String builds a Tag object into a string i.e. `json:"foo,omitempty" envconfig:"FOO"`
// note that WriteString() always returns a nil error so there is no point catching it
func (t Tag) String() string {
	var builder strings.Builder

	builder.WriteString("`")

	idx := 0
	for k, values := range t {
		builder.WriteString(k)
		builder.WriteString(":\"")
		builder.WriteString(strings.Join(values, ","))
		builder.WriteString("\"")

		if idx < len(t)-1 {
			builder.WriteString(" ")
		}
		idx++
	}

	builder.WriteString("`")
	return builder.String()
}
