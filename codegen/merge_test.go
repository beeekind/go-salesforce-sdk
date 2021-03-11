package codegen_test

import (
	"sort"
	"testing"

	"github.com/beeekind/go-salesforce-sdk/codegen"
	"github.com/stretchr/testify/require"
)

type mergeStructsInput struct {
	Structs codegen.Structs
}

type mergeStructsOutput struct {
	description string
	Results     codegen.Structs
}

var mergeTests = map[*mergeStructsInput]*mergeStructsOutput{
	{codegen.Structs{
		{
			Name:          "Banana",
			Documentation: "",
			Properties: []*codegen.Property{
				{
					Name:          "Title",
					Documentation: "",
					Type:          "int",
				},
			},
		},
		{
			Name:          "Apple",
			Documentation: "This is an apple",
		},
		{
			Name:          "Banana",
			Documentation: "A Bannana is a fruit",
			Properties: []*codegen.Property{
				{
					Name:          "Title",
					Documentation: "Title is a title",
					Type:          "int",
				},
			},
		},
	}}: {"removed duplicate structs", codegen.Structs{
		{
			Name:          "Apple",
			Documentation: "This is an apple",
		},
		{
			Name:          "Banana",
			Documentation: "A Bannana is a fruit",
			Properties: []*codegen.Property{
				{
					Name:          "Title",
					Documentation: "Title is a title",
					Type:          "int",
				},
			},
		},
	}},
	{codegen.Structs{
		{
			Name:          "Apple",
			Documentation: "This is an apple",
		},
		{
			Name:          "Banana",
			Documentation: "A Bannana is a fruit",
			Properties: codegen.Properties{
				{
					Name:          "Title",
					Documentation: "Title is a title",
					Type:          "int",
				},
			},
		},
		{
			Name:          "Banana",
			Documentation: "",
			Properties: codegen.Properties{
				{
					Name:          "Title",
					Documentation: "",
					Type:          "int",
				},
			},
		},
		{
			Name:          "Apple",
			Documentation: "",
		},
	}}: {"does not remove documentation", codegen.Structs{
		{
			Name:          "Apple",
			Documentation: "This is an apple",
		},
		{
			Name:          "Banana",
			Documentation: "A Bannana is a fruit",
			Properties: codegen.Properties{
				{
					Name:          "Title",
					Documentation: "Title is a title",
					Type:          "int",
				},
			},
		},
	}},
}

func TestMergeStructs(t *testing.T) {
	for in, out := range mergeTests {
		t.Run(out.description, func(t *testing.T) {
			results := in.Structs.Merge(nil, codegen.MergeAll, true)
			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})

			require.Equal(t, len(results), len(out.Results))

			for i := 0; i < len(out.Results); i++ {
				actual := results[i]
				expected := out.Results[i]

				require.Equal(t, expected.Name, actual.Name)
				require.Equal(t, expected.ParentName, actual.ParentName)
				require.Equal(t, expected.Documentation, actual.Documentation)
				require.Equal(t, expected.DocumentationURL, actual.DocumentationURL)
			}
		})
	}
}

type mergePropertyInput struct {
	old *codegen.Property
	new *codegen.Property
}

type mergePropertyOutput struct {
	description string
	final       *codegen.Property
}

var mergePropertyTests = map[*mergePropertyInput]*mergePropertyOutput{
	{
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "string",
			Documentation: "",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "string",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{},
		},
	}: {
		"simple merge of documentation and rag",
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "string",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
	},
	{
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "string",
			Documentation: "",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{"": []string{}},
		},
	}: {
		"merge of type and omission of empty tag",
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "string",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
	},
	{
		&codegen.Property{
			ParentName:    "Spot",
			Name:          "Rover",
			Type:          "string",
			Documentation: "",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
		&codegen.Property{
			ParentName:    "",
			Name:          "Rover",
			Type:          "",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{"": []string{}},
		},
	}: {
		"merges parent name",
		&codegen.Property{
			ParentName:    "Spot",
			Name:          "Rover",
			Type:          "string",
			Documentation: "Rover is a string",
			Tag:           codegen.Tag{"json": []string{"rover"}},
		},
	},
}

func TestMergeProperty(t *testing.T) {
	for in, out := range mergePropertyTests {
		t.Run(out.description, func(t *testing.T) {
			result := codegen.MergeProperty(*in.old, *in.new)
			require.Equal(t, *out.final, result)
		})
	}
}
