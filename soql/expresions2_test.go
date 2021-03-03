package soql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type escapeSingleQuoteInput struct {
	input string 
}

type escapeSingleQuoteOutput struct {
	output string 
	description string 
}

var escapeSingleQuoteTests = map[*escapeSingleQuoteInput]*escapeSingleQuoteOutput{
	{"bar' AND zar = 'lar"}:{`bar\' AND zar = \'lar`, "escape an addition to the where clause"},
	{`bar\' AND zar =\'lar`}:{`bar\' AND zar =\'lar`, "does not escape a string that is already escaped"},
}

func TestEscapeSingleQuote(t *testing.T) {
	for in, out := range escapeSingleQuoteTests {
		t.Run(out.description, func(t *testing.T) {
			result := escapeSingleQuote(in.input)
			require.Equal(t, out.output, result)
		})
	}
}
