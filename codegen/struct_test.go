package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type jsonTagInput struct {
	values []string
}

type jsonTagOutput struct {
	description string
	tag         Tag
}

var jsonTagTests = map[*jsonTagInput]*jsonTagOutput{
	{[]string{"foo"}}: {"basic json tag", Tag{
		"json": []string{"foo"},
	}},
	{[]string{"bar", "omitempty"}}: {"include secondary omitempty tag", Tag{
		"json": []string{"bar", "omitempty"},
	}},
}

func TestJSONTag(t *testing.T) {
	for in, out := range jsonTagTests {
		t.Run(out.description, func(t *testing.T) {
			result := jsonTag(in.values...)
			require.Equal(t, out.tag, result)
		})
	}
}

type tagStringInput struct {
	tag Tag
}

type tagStringOutput struct {
	description string
	output      string
}

var tagStringTests = map[*tagStringInput]*tagStringOutput{
	{Tag{
		"json": []string{"foo"},
	}}: {"basic json tag", "`json:\"foo\"`"},
	{Tag{
		"json": []string{"bar", "omitempty"},
	}}: {"tag + omitempty", "`json:\"bar,omitempty\"`"},
	{Tag{
		"json": []string{"bar", "zar", "omitempty"},
	}}: {"tag + omitempty", "`json:\"bar,zar,omitempty\"`"},
}

func TestTagString(t *testing.T) {
	for in, out := range tagStringTests {
		t.Run(out.description, func(t *testing.T) {
			result := in.tag.String()
			require.Equal(t, out.output, result)
		})
	}
}