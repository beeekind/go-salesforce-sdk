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
					Type: "interface{}",
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
