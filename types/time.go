package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DefaultDateFormats ...
// first element is used as the string representation for time.Format()
var DefaultDateFormats = []string{
	"2006-01-02",
	"01/02/2006",
}

const (
	// ISODate ...
	ISODate = "2006-01-02"
)

// DefaultDatetimeFormats ...
// first element is used as the string representation for time.Format()
var DefaultDatetimeFormats = []string{
	"2006-01-02T15:04:05.000-0700",
	//"2006-01-02T15:04:05.999Z",
	//"2006-01-02 15:04:05",
	//"01/02/2006 15:04:05",
}

// Date ...
type Date struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      time.Time
}

// NewDate ...
func NewDate(t time.Time) Date {
	return Date{
		IsNull: false,
		Value:  t,
	}
}

// ParseDate ...
func ParseDate(str string, formats ...string) (Date, error) {
	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, str)
		if err == nil {
			return NewDate(t), nil
		}
	}

	return Date{}, fmt.Errorf("ParseDate: %s: %w", str, err)
}

// MarshalJSON ...
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsNull {
		return []byte("null"), nil
	}

	if d.Value.IsZero() {
		return nil, errors.New("attempted to marshal a 0-value time.Time - I don't think 0001 - 01 - 01 is a date you want to work with")
	}

	return []byte(d.Value.Format(ISODate)), nil
}

// UnmarshalJSON ...
func (d *Date) UnmarshalJSON(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		d.IsNull = true
		return nil
	}

	t, err := ParseDate(string(data), DefaultDateFormats...)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON date: %w", err)
	}

	d.Value = t.Value
	d.IsNull = false
	d.IsHydrated = true
	return nil
}

// MarshalText ...
func (d Date) MarshalText() ([]byte, error) {
	if d.IsNull {
		return []byte("null"), nil
	}

	return []byte(d.Value.Format(DefaultDateFormats[0])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (d *Date) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		d.IsNull = true
		return nil
	}

	t, err := ParseDate(string(data))
	if err != nil {
		return fmt.Errorf("UnmarshalingText date: %w", err)
	}

	d.Value = t.Value
	d.IsNull = false
	d.IsHydrated = true
	return nil
}

func (d Date) String() string {
	return d.Value.Format(DefaultDateFormats[0])
}

// Datetime ...
type Datetime struct {
	IsNull     bool `json:"-"`
	IsHydrated bool `json:"-"`
	Value      time.Time
}

// NewDatetime ...
func NewDatetime(t time.Time) Datetime {
	return Datetime{
		IsNull: false,
		Value:  t,
	}
}

// ParseDatetime ...
func ParseDatetime(str string, formats ...string) (Datetime, error) {
	var t time.Time
	var err error
	str = strings.TrimSuffix(strings.TrimPrefix(str, `"`), `"`)
	for _, format := range formats {
		t, err = time.Parse(format, str)
		if err == nil {
			return NewDatetime(t), nil
		}
	}

	return Datetime{}, fmt.Errorf("ParseDate: %s: %w", str, err)
}

// MarshalJSON ...
func (d Datetime) MarshalJSON() ([]byte, error) {
	if d.IsNull {
		return []byte("null"), nil
	}

	return []byte(d.Value.Format(DefaultDatetimeFormats[0])), nil
}

// UnmarshalJSON ...
func (d *Datetime) UnmarshalJSON(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		d.IsNull = true
		return nil
	}

	t, err := ParseDatetime(string(data), DefaultDatetimeFormats...)
	if err != nil {
		println(5, string(data))
		return fmt.Errorf("UnmarshalJSON types.Datetime: %w", err)
	}

	//println(3, string(data))
	//println(4, t.String())
	d.Value = t.Value
	d.IsNull = false
	d.IsHydrated = true 
	return nil
}

// MarshalText ...
func (d Datetime) MarshalText() ([]byte, error) {
	if d.IsNull {
		return []byte("null"), nil
	}

	return []byte(d.Value.Format(DefaultDatetimeFormats[0])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (d *Datetime) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		d.IsNull = true
		return nil
	}

	t, err := ParseDatetime(string(data))
	if err != nil {
		return fmt.Errorf("UnmarshalingText date: %w", err)
	}

	d.Value = t.Value
	d.IsNull = false
	d.IsHydrated = true 
	return nil
}

func (d Datetime) String() string {
	return d.Value.Format(DefaultDatetimeFormats[0])
}
