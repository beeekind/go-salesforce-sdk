package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// absurd.go contains absurd types that shouldn't have to exist

var badBool = []byte(`"N"`)
var goodBoolTrue = []byte(`true`)
var goodBoolFalse = []byte(`false`)

// AlmostBool ...
//
// You can't believe its not ~butter~ boolean.
//
// The JSON response type of the /sobjects/{objectName}/describe endpoint
// will return an array of objects representing a Field and the key defaultValue has 
// an innovative return type of... drum roll please... {null, true, false, "N"}
// 
// At least its documented... oh wait no its not the field isn't documented at all (02/09/2020) !
// https://developer.salesforce.com/docs/atlas.en-us.230.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm
type AlmostBool struct {
	IsNull bool
	Value  bool
}

// MarshalJSON ...
func (b *AlmostBool) MarshalJSON() ([]byte, error) {
	if b.IsNull {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatBool(b.Value)), nil
}

// UnmarshalJSON ...
func (b *AlmostBool) UnmarshalJSON(data []byte) (err error) {
	if bytes.Equal(data, nullBytes){
		b.IsNull = true 
		return nil 
	} 
	
	// some boolean properties of the salesforce rest api return string values
	if !bytes.Equal(data, goodBoolTrue) && !bytes.Equal(data, goodBoolFalse){
		b.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &b.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON for AlmostBool: %w", err)
	}

	b.IsNull = false
	return nil
}

// MarshalText ...
func (b *AlmostBool) MarshalText() ([]byte, error) {
	if b.IsNull {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatBool(b.Value)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (b *AlmostBool) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		b.IsNull = true
		return nil
	}

	var err error
	b.Value, err = strconv.ParseBool(string(data))
	if err != nil {
		return fmt.Errorf("UnmarshalText for AlmostBool: %w", err)
	}

	return nil
}

func (b *AlmostBool) String() string {
	return strconv.FormatBool(b.Value)
}
