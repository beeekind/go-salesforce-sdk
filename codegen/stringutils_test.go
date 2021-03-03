package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type convertInitialismInput struct {
	input string
}

type convertInitialismOutput struct {
	description string
	output      string
}

var initialismTests = map[*convertInitialismInput]*convertInitialismOutput{
	{"someJson"}: {"someJson", "someJSON"},
	{"someApi"}:  {"someApi", "someAPI"},
}

func TestInitialisms(t *testing.T) {
	for in, out := range initialismTests {
		t.Run(out.description, func(t *testing.T) {
			output := convertInitialisms(in.input)
			require.Equal(t, out.output, output)
		})
	}
}

type camelInput struct {
	input    string
	initCase bool
}

type camelOutput struct {
	description string
	output      string
}

var camelCaseTests = map[*camelInput]*camelOutput{
	// toUpperCamelCase
	{"test_case", true}:            {"test_case", "TestCase"},
	{"test.case", true}:            {"test.case", "TestCase"},
	{"test", true}:                 {"test", "Test"},
	{"TestCase", true}:             {"TestCase", "TestCase"},
	{" test  case ", true}:         {" test  case ", "TestCase"},
	{"", true}:                     {"", ""},
	{"many_many_words", true}:      {"many_many_words", "ManyManyWords"},
	{"AnyKind of_string", true}:    {"AnyKind of_string", "AnyKindOfString"},
	{"odd-fix", true}:              {"odd-fix", "OddFix"},
	{"numbers2And55with000", true}: {"numbers2And55With000", "Numbers2And55With000"},
	// toLowerCamelCase
	{"test_case", false}:            {"test_case", "testCase"},
	{"test.case", false}:            {"test.case", "testCase"},
	{"test", false}:                 {"test", "test"},
	{"TestCase", false}:             {"TestCase", "testCase"},
	{" test  case ", false}:         {" test  case ", "testCase"},
	{"", false}:                     {"", ""},
	{"many_many_words", false}:      {"many_many_words", "manyManyWords"},
	{"AnyKind of_string", false}:    {"AnyKind of_string", "anyKindOfString"},
	{"odd-fix", false}:              {"odd-fix", "oddFix"},
	{"numbers2And55with000", false}: {"numbers2And55With000", "numbers2And55With000"},
}

func TestCamelCasing(t *testing.T) {
	for in, out := range camelCaseTests {
		t.Run(out.description, func(t *testing.T) {
			result := toCamelInitCase(in.input, in.initCase)
			require.Equal(t, out.output, result)
		})
	}
}

type convertTypeInput struct {
	input string
}

type convertTypeOutput struct {
	description string
	result      string
	err         error
}

var convertTypeTests = map[*convertTypeInput]*convertTypeOutput{
	{"reference"}:                          {"reference", "string", nil},
	{"address"}:                            {"address", "types.Address", nil},
	{"currency"}:                           {"currency", "float64", nil},
	{"string"}:                             {"string", "string", nil},
	{"picklist"}:                           {"picklist", "string", nil},
	{"date"}:                               {"date", "types.Date", nil},
	{"text"}:                               {"text", "string", nil},
	{"number"}:                             {"number", "int", nil},
	{"lookup"}:                             {"lookup", "string", nil},
	{"textarea"}:                           {"textarea", "string", nil},
	{"roll-up summary (sum invoice line)"}: {"roll-up", "string", nil},
	{"dateTime"}:                           {"dateTime", "types.Datetime", nil},
	{"email"}:                              {"email", "string", nil},
	{"phone"}:                              {"phone", "string", nil},
	{"datetime"}:                           {"datetime", "types.Datetime", nil},
	{"boolean"}:                            {"boolean", "bool", nil},
	{"double"}:                             {"double", "float64", nil},
	{"url"}:                                {"url", "string", nil},
	{"id"}:                                 {"id", "string", nil},
	{"anyType"}:                            {"anyType", "json.RawMessage", nil},
	{"decimal"}:                            {"decimal", "float64", nil},
	{"long"}:                               {"long", "int64", nil},
	{"object"}:                             {"object", "json.RawMessage", nil},
	{"phone"}:                              {"phone", "string", nil},
	{"percent"}:                            {"percent", "float64", nil},
	{"date/time"}:                          {"date/time", "types.Datetime", nil},
}

func TestConvertType(t *testing.T) {
	for in, out := range convertTypeTests {
		t.Run(out.description, func(t *testing.T) {
			result, err := convertType(in.input)
			require.Equal(t, out.err, err)
			require.Equal(t, out.result, result)
		})
	}
}

type convertNullableInput struct {
	input string
}

type convertNullableOutput struct {
	description string
	result      string
	err         error
}

var convertNullabletests = map[*convertNullableInput]*convertNullableOutput{}

func TestConvertNullable(t *testing.T) {
	for in, out := range convertNullabletests {
		t.Run(out.description, func(t *testing.T) {
			result, err := convertNillableType(in.input)
			require.Equal(t, out.err, err)
			require.Equal(t, out.result, result)
		})
	}
}
