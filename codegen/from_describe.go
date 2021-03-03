package codegen

// from_description.go converts a /sobjects/{objectName}/describe response into codegen.Struct
// and codegen.Property objects that can be used to generate go code

import (
	"bytes"
	"errors"
	"sort"
	"strings"

	"github.com/b3ntly/salesforce/metadata"
)

// FromDescribe converts the result of a /sobjects/{objectName}/describe request to a list of objects
// representing golang struct types annotated with their parent and child dependencies 
func FromDescribe(desc *metadata.Describe, ignoreRelations bool) (structs Structs, err error) {
	structName, err := prepareStructName(desc.Name, neither)
	if err != nil {
		return nil, err
	}

	structs = []*Struct{
		{Name: structName},
	}

	for _, f := range desc.Fields {
		if !ignoreRelations && isForeignKey(f) {
			relatedProperties, dependencies, polymorphicModel, err := convertForeignKey(structName, f)
			if err != nil {
				return nil, err
			}

			if polymorphicModel != nil {
				structs = append(structs, polymorphicModel)
			}

			structs[0].Dependencies = append(structs[0].Dependencies, dependencies...)
			structs[0].Properties = append(structs[0].Properties, relatedProperties...)
			continue
		}

		prop, err := convertField(f)
		if err != nil {
			return nil, err
		}

		structs[0].Properties = append(structs[0].Properties, prop)
	}

	if !ignoreRelations {
		for _, ch := range desc.ChildRelationships {
			// ignore child relationships without established relationship names - it's
			// unclear why some child relationships wouldn't have relationship names. It could
			// be field-level security, or some kind of internal alias that isn't intended to be
			// queried.
			if ch.RelationshipName == "" {
				continue
			}

			prop, dependencies, polymorphicModel, err := convertChild(structName, ch)
			if err != nil {
				return nil, err
			}

			if polymorphicModel != nil {
				structs = append(structs, polymorphicModel)
			}

			structs[0].Dependencies = append(structs[0].Dependencies, dependencies...)
			structs[0].Properties = append(structs[0].Properties, prop)
		}
	}

	// do not include this object as a dependency of itself 
	structs[0].Dependencies = dedupe(structs[0].Dependencies, structs[0].Name)
	if !structs.Valid() {
		return nil, errors.New("codegen.FromDescribe produced invalid structs")
	}

	return structs, nil
}

func isForeignKey(f *metadata.Field) bool {
	return len(f.ReferenceTo) > 0 && f.RelationshipName != ""
}

// a foreignKey property is represented by two Property objects on the underlying codegen.Struct.
// 1) the foreign key itself such as WhoId which is a Salesforce ID string
// 2) the response type representing the fields of the related object returned by a query for example,
//    Event.Who.Name returned from select Who.Name from Events where WhoId = 'somesalesforceid'
//
//    Resulting in:
//
//	  type Event struct {
//	        WhoID string `json:"WhoId"`
//	  		Who   *LeadAccount `json:"Who"`
//	  }
//
//    type LeadAccount struct {
//        Lead
//        Account
//    }
//
// Returns:
// {props} the two direct properties resulting from this foreign key
// {dependencies} a slice of strings representing non-polymorphic objects the props depend on
// {polymorphicModel} a struct (or nil) that the props depend on. If the field is not a polymorphic key the value will be nil.
// {error}
func convertForeignKey(entityName string, f *metadata.Field) (props []*Property, dependencies []string, polymorphicModel *Struct, err error) {
	name, err := prepareStructName(f.Name, singular)
	if err != nil {
		return nil, nil, nil, err
	}

	if f.RelationshipName == "" {
		return nil, nil, nil, errors.New("foreign key property is missing a relationshipName field")
	}

	foreignKey := &Property{
		Name: name,
		Type: "string",
		Tag:  jsonTag(f.Name),
	}

	polymorphicType := convertPolymorphicKey(f.ReferenceTo)
	relatedObject := &Property{
		ParentName: entityName,
		Name:       f.RelationshipName,
		Type:       "*" + polymorphicType,
		Tag:        jsonTag(f.RelationshipName),
	}

	polymorphicModel = convertPolymorphicFields(f.ReferenceTo)
	return []*Property{foreignKey, relatedObject}, f.ReferenceTo, polymorphicModel, nil
}

func convertField(f *metadata.Field) (prop *Property, err error) {
	prop = &Property{
		Name:       convertInitialisms(f.Name),
		Tag:        jsonTag(f.Name),
		IsNillable: f.Nillable,
	}
	prop.Type, err = convertType(f.Type)
	if err != nil {
		return prop, err
	}
	return prop, nil
}

func convertChild(entityName string, ch *metadata.ChildRelationship) (prop *Property, dependencies []string, polymorphicModel *Struct, err error) {
	prop = &Property{
		ParentName: entityName,
		Name:       ch.RelationshipName,
		Tag:        jsonTag(ch.RelationshipName),
	}

	dType := ch.ChildSObject
	var strs []string
	if len(ch.JunctionReferenceTo) > 0 {
		for _, s := range ch.JunctionReferenceTo {
			strs = append(strs, string(s))
		}
		dType = convertPolymorphicKey(strs)
		dependencies = append(dependencies, strs...)
	} else {
		if ch.ChildSObject != "" {
			dependencies = append(dependencies, ch.ChildSObject)
		}
	}

	var buff bytes.Buffer
	if err := childTypeTemplate.Execute(&buff, dType); err != nil {
		return prop, nil, nil, err
	}

	prop.Type = buff.String()
	polymorphicModel = convertPolymorphicFields(strs)
	return prop, dependencies, polymorphicModel, nil
}

// a polymorphic key is a foreign key that may reference a varierty of different salesforce objects
func convertPolymorphicKey(referenceTos []string) string {
	if len(referenceTos) == 1 {
		return referenceTos[0]
	}

	// I'm unsure if the slice ordering is the same every time so to create a consistent
	// parent name sort this slice alphabetically
	sort.Slice(referenceTos, func(i, j int) bool {
		return referenceTos[i] < referenceTos[j]
	})

	// []string{"Lead", "Account"} => "LeadAccount"
	return strings.Join(referenceTos, "")
}

func convertPolymorphicFields(referenceTos []string) *Struct {
	// if there are not multiple references by a given field it is not polymorphic
	if len(referenceTos) < 2 {
		return nil
	}

	// compute the name first as this will also sort the referenceTos array
	// and produce alphabetically sorted dependantProperties
	name := convertPolymorphicKey(referenceTos)

	var dependantProperties []*Property
	for _, referenceTo := range referenceTos {
		dependantProperties = append(dependantProperties, &Property{
			Name:       referenceTo,
			IsEmbedded: true,
		})
	}

	return &Struct{
		Name:               name,
		Properties:         dependantProperties,
		IsPolymorphicModel: true,
	}
}
