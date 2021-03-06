package codegen_test

import (
	"sort"
	"testing"

	"github.com/beeekind/go-salesforce-sdk/codegen"
	"github.com/stretchr/testify/require"
)

type json2goInput struct {
	name          string
	documentation string
	payload       []byte
}

type json2goOutput struct {
	description string
	results     []codegen.Struct
	err         error
}

var json2goTests = map[*json2goInput]json2goOutput{
	{"Foo", "", []byte(`{
		"foo": "bar",
		"bar": 4,
		"see": 0.432,
		"lar": ["nar", "sar"],
		"zar": {"lar": "tar"},
		"nar": [33, 22],
		"fields": [
			{"par": 4}
		],
		"tar": {
			"foo": "bar",
			"bar": 4,
			"see": 0.432,
			"lar": ["nar", "sar"],
			"zar": {"lar": "tar"},
			"nar": [33, 22],
			"fields": [
				{"par": 4}
			]
		}
	}`)}: {"basic object", []codegen.Struct{
		{
			Name: "Foo",
			Properties: codegen.Properties{
				{"", "Foo", "", "string", codegen.Tag{"json": []string{"foo"}}, false, false},
				{"", "Bar", "", "int64", codegen.Tag{"json": []string{"bar"}}, false, false},
				{"", "See", "", "float64", codegen.Tag{"json": []string{"see"}}, false, false},
				{"", "Lar", "", "[]string", codegen.Tag{"json": []string{"lar"}}, false, false},
				{"", "Zar", "", "*Zar", codegen.Tag{"json": []string{"zar"}}, false, false},
				{"", "Nar", "", "[]int64", codegen.Tag{"json": []string{"nar"}}, false, false},
				{"", "Fields", "", "[]*Field", codegen.Tag{"json": []string{"fields"}}, false, false},
				{"", "Tar", "", "*Tar", codegen.Tag{"json": []string{"tar"}}, false, false},
			},
		},
		{
			Name: "Tar",
			Properties: codegen.Properties{
				{"", "Foo", "", "string", codegen.Tag{"json": []string{"foo"}}, false, false},
				{"", "Bar", "", "int64", codegen.Tag{"json": []string{"bar"}}, false, false},
				{"", "See", "", "float64", codegen.Tag{"json": []string{"see"}}, false, false},
				{"", "Lar", "", "[]string", codegen.Tag{"json": []string{"lar"}}, false, false},
				{"", "Zar", "", "*Zar", codegen.Tag{"json": []string{"zar"}}, false, false},
				{"", "Nar", "", "[]int64", codegen.Tag{"json": []string{"nar"}}, false, false},
				{"", "Fields", "", "[]*Field", codegen.Tag{"json": []string{"fields"}}, false, false},
			},
		},
		{
			Name:       "Field",
			ParentName: "Foo",
			Properties: codegen.Properties{
				{"", "Par", "", "int64", codegen.Tag{"json": []string{"par"}}, false, false},
			},
		},
		{
			Name:       "Field",
			ParentName: "Tar",
			Properties: codegen.Properties{
				{"", "Par", "", "int64", codegen.Tag{"json": []string{"par"}}, false, false},
			},
		},
		{
			Name:       "Zar",
			ParentName: "Foo",
			Properties: codegen.Properties{
				{"", "Lar", "", "string", codegen.Tag{"json": []string{"lar"}}, false, false},
			},
		},
		{
			Name:       "Zar",
			ParentName: "Tar",
			Properties: codegen.Properties{
				{"", "Lar", "", "string", codegen.Tag{"json": []string{"lar"}}, false, false},
			},
		},
	}, nil},
}

func TestCodegenJSON2Go(t *testing.T) {
	for in, out := range json2goTests {
		t.Run(out.description, func(t *testing.T) {
			results, err := codegen.FromJSON(in.name, in.documentation, in.payload)
			require.Equal(t, out.err, err)
			require.Equal(t, len(out.results), len(results))

			sort.Slice(out.results, func(i, j int) bool {
				return out.results[i].Name < out.results[j].Name
			})

			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})

			for idx, result := range results {
				require.Equal(t, out.results[idx].Name, result.Name)

				sort.Slice(result.Properties, func(i, j int) bool {
					return result.Properties[i].Name < result.Properties[j].Name
				})

				sort.Slice(out.results[idx].Properties, func(i, j int) bool {
					return out.results[idx].Properties[i].Name < out.results[idx].Properties[j].Name
				})

				for idx2, prop := range result.Properties {
					//println("======")
					//println(result.Name, prop.Name)
					//println(out.results[idx].Name, out.results[idx].Properties[idx2].Name)
					require.Equal(t, out.results[idx].Properties[idx2], prop)
				}
			}
		})
	}
}

type structStringInput json2goInput

type structStringOutput struct {
	description string
	err         error
	code        []string
}

var structStringTests = map[*structStringInput]*structStringOutput{
	{"Foo", "", []byte(`{
		"foo": "bar",
		"bar": 4,
		"see": 0.432,
		"lar": ["nar", "sar"],
		"zar": {"lar": "tar"},
		"nar": [33, 22],
		"fields": [
			{"par": 4}
		],
		"tar": {
			"foo": "bar",
			"bar": 4,
			"see": 0.432,
			"lar": ["nar", "sar"],
			"zar": {"lar": "tar"},
			"nar": [33, 22],
			"fields": [
				{"par": 4}
			]
		}
	}`)}: {"FooExample", nil, []string{
		"\n// Field ... \ntype Field struct {\n\t// Par ... \n\tPar int64 `json:\"par\"`\n}",

		"\n// Field ... \ntype Field struct {\n\t// Par ... \n\tPar int64 `json:\"par\"`\n}",

		"\n// Foo ... \ntype Foo struct {\n\t// Bar ... \n\tBar int64 `json:\"bar\"`\n\t// Fields ... \n\tFields []*Field `json:\"fields\"`\n\t// Foo ... \n\tFoo string `json:\"foo\"`\n\t// Lar ... \n\tLar []string `json:\"lar\"`\n\t// Nar ... \n\tNar []int64 `json:\"nar\"`\n\t// See ... \n\tSee float64 `json:\"see\"`\n\t// Tar ... \n\tTar *Tar `json:\"tar\"`\n\t// Zar ... \n\tZar *Zar `json:\"zar\"`\n}",

		"\n// Tar ... \ntype Tar struct {\n\t// Bar ... \n\tBar int64 `json:\"bar\"`\n\t// Fields ... \n\tFields []*Field `json:\"fields\"`\n\t// Foo ... \n\tFoo string `json:\"foo\"`\n\t// Lar ... \n\tLar []string `json:\"lar\"`\n\t// Nar ... \n\tNar []int64 `json:\"nar\"`\n\t// See ... \n\tSee float64 `json:\"see\"`\n\t// Zar ... \n\tZar *Zar `json:\"zar\"`\n}",

		"\n// Zar ... \ntype Zar struct {\n\t// Lar ... \n\tLar string `json:\"lar\"`\n}",

		"\n// Zar ... \ntype Zar struct {\n\t// Lar ... \n\tLar string `json:\"lar\"`\n}",
	}},
}

func TestStructString(t *testing.T) {
	for in, out := range structStringTests {
		t.Run(out.description, func(t *testing.T) {
			results, err := codegen.FromJSON(in.name, in.documentation, in.payload)
			require.Equal(t, out.err, err)

			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})

			for idx, result := range results {
				src, err := result.String()
				if err != nil {
					t.Log(idx, err.Error())
				}

				require.Nil(t, err)
				require.Equal(t, out.code[idx], src, idx)
			}
		})
	}
}

type newPropertyInput struct {
	parent        string
	name          string
	dataType      string
	documentation string
	tagKey        string
	tagValues     []string
}

type newPropertyOutput struct {
	description string
	result      codegen.Property
}

var newPropertyTests = map[*newPropertyInput]*newPropertyOutput{
	{"", "Title", "text", "", "json", []string{"title"}}: {"", codegen.Property{
		"",
		"Title",
		"",
		"string",
		codegen.Tag{"json": []string{"title"}},
		false,
		false,
	}},
	{"", "ParentId", "reference", "", "json", []string{"ParentId"}}: {"", codegen.Property{
		"",
		"ParentID",
		"",
		"string",
		codegen.Tag{"json": []string{"ParentId"}},
		false,
		false,
	}},
}

func TestNewProperty(t *testing.T) {
	for in, out := range newPropertyTests {
		t.Run(out.description, func(t *testing.T) {
			result, err := codegen.NewProperty(in.parent, in.name, in.dataType, in.documentation, in.tagKey, in.tagValues...)
			require.Nil(t, err)
			require.Equal(t, out.result, result)
		})
	}
}

type newRelationPropertyInput struct {
	parentName   string
	relationName string
}

type newRelationPropertyOutput struct {
	description string
	result      *codegen.Property
}

var newRelationPropertyTests = map[*newRelationPropertyInput]*newRelationPropertyOutput{
	{"Lead", "Leads"}: {"", &codegen.Property{
		ParentName:    "Lead",
		Name:          "Leads",
		Documentation: "",
		Type:          "struct {\n\tDone bool `json:\"done\"`\n\tCount int `json:\"count\"`\n\tTotalSize int `json:\"totalSize\"`\n\tRecords []*Lead `json:\"records\"`\n}",
		Tag:           codegen.Tag{"json": []string{"Leads"}},
		IsEmbedded:    false,
	}},
	{"Contact", "Employees"}: {"", &codegen.Property{
		ParentName:    "Contact",
		Name:          "Employees",
		Documentation: "",
		Type:          "struct {\n\tDone bool `json:\"done\"`\n\tCount int `json:\"count\"`\n\tTotalSize int `json:\"totalSize\"`\n\tRecords []*Contact `json:\"records\"`\n}",
		Tag:           codegen.Tag{"json": []string{"Employees"}},
		IsEmbedded:    false,
	}},
}

func TestNewRelationshipProperty(t *testing.T) {
	for in, out := range newRelationPropertyTests {
		t.Run(out.description, func(t *testing.T) {
			result, err := codegen.NewChildProperty(in.parentName, in.relationName)
			require.Nil(t, err)
			require.Equal(t, out.result, result)
		})
	}
}

type overlapInput struct {
	name     string
	contents [][]byte
}

type overlapOutput struct {
	outputs codegen.Structs
}

var overlapTests = map[*overlapInput]*overlapOutput{
	{"Foo", [][]byte{[]byte(`{"bar": null}`), []byte(`{"bar": ["one", "two"]}`)}}: {
		codegen.Structs{
			{Name: "Foo", Properties: []*codegen.Property{
				{
					Name: "Bar",
					Type: "[]string",
					Tag:  codegen.Tag{"json": []string{"bar"}},
				},
			}},
		},
	},
}

func TestProperyOverlaps(t *testing.T) {
	for in, out := range overlapTests {
		t.Run(in.name, func(t *testing.T) {
			var results codegen.Structs

			for _, c := range in.contents {
				structs, err := codegen.FromJSON(in.name, "", c)
				require.Nil(t, err)
				results = append(results, structs...)
			}

			results = results.Dedupe(true)
			require.Equal(t, len(out.outputs), len(results))

			for i, o := range out.outputs {
				for j, p := range o.Properties {
					p2 := results[i].Properties[j]
					require.Equal(t, p.Name, p2.Name)
					require.Equal(t, p.Type, p2.Type)
					require.Equal(t, p.Tag, p2.Tag)
				}
			}
		})
	}
}
