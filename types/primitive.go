package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

var nullBytes = []byte("null")

// NullableBool ...
type NullableBool struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      bool
}

// MarshalJSON ...
func (b *NullableBool) MarshalJSON() ([]byte, error) {
	if b.IsNull {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatBool(b.Value)), nil
}

// UnmarshalJSON ...
func (b *NullableBool) UnmarshalJSON(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		b.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &b.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON for NullableBool: %w", err)
	}

	b.IsNull = false
	b.IsHydrated = true
	return nil
}

// MarshalText ...
func (b *NullableBool) MarshalText() ([]byte, error) {
	if b.IsNull {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatBool(b.Value)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (b *NullableBool) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		b.IsNull = true
		return nil
	}

	var err error
	b.Value, err = strconv.ParseBool(string(data))
	if err != nil {
		return fmt.Errorf("UnmarshalText for NullableBool: %w", err)
	}

	b.IsHydrated = true
	return nil
}

func (b *NullableBool) String() string {
	return strconv.FormatBool(b.Value)
}

// NullableString ...
type NullableString struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      string
}

// NewString returns a new instance of NullableString
func NewString(str string) NullableString {
	return NullableString{
		IsNull: false,
		Value:  str,
	}
}

// MarshalJSON converts NullableString into an appropriate JSON representation
func (s NullableString) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON decodes json data into a NullableString
func (s *NullableString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON for NullableString: %w", err)
	}

	s.IsNull = false
	s.IsHydrated = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a null representation when this String is null.
func (s *NullableString) MarshalText() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return []byte(s.Value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullableString) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	s.Value = string(data)
	s.IsNull = false
	s.IsHydrated = true
	return nil
}

// String ...
func (s *NullableString) String() string {
	if s.IsNull {
		return ""
	}

	return s.Value
}

// NullableInt ...
type NullableInt struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      int
}

// NewInt ...
func NewInt(val int) NullableInt {
	return NullableInt{
		IsNull: false,
		Value:  val,
	}
}

// MarshalJSON converts NullableString into an appropriate JSON representation
func (s NullableInt) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON decodes json data into a NullableString
func (s *NullableInt) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON for NullableInt: %w", err)
	}

	s.IsNull = false
	s.IsHydrated = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a null representation when this String is null.
func (s *NullableInt) MarshalText() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return []byte(strconv.Itoa(s.Value)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullableInt) UnmarshalText(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	s.Value, err = strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("UnmarshalText for NullableInt: %w", err)
	}
	s.IsNull = false
	s.IsHydrated = true
	return nil
}

// String ...
func (s *NullableInt) String() string {
	if s.IsNull {
		return ""
	}

	return strconv.Itoa(s.Value)
}

// NullableInt64 ...
type NullableInt64 struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      int64
}

// NewInt64 ...
func NewInt64(val int64) NullableInt64 {
	return NullableInt64{
		IsNull: false,
		Value:  val,
	}
}

// MarshalJSON converts NullableString into an appropriate JSON representation
func (s *NullableInt64) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON decodes json data into a NullableString
func (s *NullableInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON NullableInt64: %w", err)
	}

	s.IsNull = false
	s.IsHydrated = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a null representation when this String is null.
func (s *NullableInt64) MarshalText() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(s.Value, 10)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullableInt64) UnmarshalText(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	s.Value, err = strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("UnmarshalText for NullableInt64: %w", err)
	}
	s.IsNull = false
	s.IsHydrated = true 
	return nil
}

// String ...
func (s *NullableInt64) String() string {
	if s.IsNull {
		return ""
	}

	return strconv.FormatInt(s.Value, 10)
}

// NullableFloat64 ...
type NullableFloat64 struct {
	IsNull bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value  float64
}

// NewFloat64 ...
func NewFloat64(val float64) NullableFloat64 {
	return NullableFloat64{
		IsNull: false,
		Value:  val,
	}
}

// MarshalJSON converts NullableString into an appropriate JSON representation
func (s *NullableFloat64) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(s.Value)
}

// UnmarshalJSON decodes json data into a NullableString
func (s *NullableFloat64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.Value); err != nil {
		return fmt.Errorf("UnmarshalJSON for NullableFloat64: %w", err)
	}

	s.IsNull = false
	s.IsHydrated = true 
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a null representation when this String is null.
func (s *NullableFloat64) MarshalText() ([]byte, error) {
	if s.IsNull {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatFloat(s.Value, 'E', -1, 64)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullableFloat64) UnmarshalText(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	s.Value, err = strconv.ParseFloat(string(data), 64)
	if err != nil {
		return fmt.Errorf("UnmarshalText for NullableInt64: %w", err)
	}
	s.IsNull = false
	s.IsHydrated = true 
	return nil
}

// String ...
func (s *NullableFloat64) String() string {
	if s.IsNull {
		return ""
	}

	return strconv.FormatFloat(s.Value, 'E', -1, 64)
}
