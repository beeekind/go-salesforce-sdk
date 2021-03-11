package requests_test

import (
	"testing"

	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/stretchr/testify/require"
)

type mergeJSONArraysInput struct {
	inputs [][]byte
}

type mergeJSONArraysOutput struct {
	output []byte
}

var mergeJSONArraysTests = map[*mergeJSONArraysInput]mergeJSONArraysOutput{
	{[][]byte{
		[]byte("[{\"a\": 1}]"),
		[]byte("[{\"a\": 1}]"),
		[]byte("[{\"b\": 1}]"),
	}}: {
		[]byte("[{\"a\": 1},{\"a\": 1},{\"b\": 1}]"),
	},
}

func TestMergeJSONArrays(t *testing.T) {
	for in, out := range mergeJSONArraysTests {
		result := requests.MergeJSONArrays(in.inputs...)
		require.Equal(t, result, out.output)
	}
}
