package types

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Attributes are metadata properties used in certain APIs such as the composite API
type Attributes struct {
	Type        string `json:"type"`
	ReferenceID string `json:"referenceId"`
}

// QueryResponse is the generic result of a query the the query or tooling/query API endpoints
type QueryResponse struct {
	Count          int             `json:"count"`
	Records        json.RawMessage `json:"records"`
	TotalSize      int             `json:"totalSize,omitempty"`
	Done           bool           `json:"done,omitempty"`
	NextRecordsURL string          `json:"nextRecordsUrl"`
}

// QueryParts is the generic result of a query the the Query API endpoint
type QueryParts struct {
	Count          int               `json:"count"`
	Records        []json.RawMessage `json:"records"`
	TotalSize      int               `json:"totalSize,omitempty"`
	Done           *bool             `json:"done,omitempty"`
	NextRecordsURL string            `json:"nextRecordsUrl"`
}

// Address ...
type Address struct {
	IsNull      bool `json:"-"`
	Accuracy    string
	City        string
	Country     string
	CountryCode string
	Latitude    float64
	Longitude   float64
	PostalCode  string
	State       string
	StateCode   string
	Street      string
}

// MarshalJSON ...
func (a *Address) MarshalJSON() ([]byte, error) {
	if a.IsNull {
		return []byte("null"), nil
	}

	// marshalling the *Address object causes an infinite recursive loop as 
	// MarshalJSON is called over and over again 
	contents, err := json.Marshal(map[string]interface{}{
		"IsNull": a.IsNull,
		"Accuracy": a.Accuracy,
		"City": a.City,
		"Country": a.Country,
		"CountryCode": a.CountryCode,
		"Latitude": a.Latitude,
		"Longitude": a.Longitude,
		"PostalCode": a.PostalCode,
		"State": a.State,
		"StateCode": a.StateCode,
		"Street": a.Street, 
	})

	if err != nil {
		return nil, err 
	}

	return contents, nil 
}

// UnmarshalJSON ...
func (a Address) UnmarshalJSON(data []byte) (err error) {
	if bytes.Equal(data, nullBytes) {
		a.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("UnmarshalJSON for NullableBool: %w", err)
	}

	a.IsNull = false
	return nil
}

// MarshalText ...
func (a Address) MarshalText() ([]byte, error) {
	if a.IsNull {
		return []byte("null"), nil
	}

	// marshalling the *Address object causes an infinite recursive loop as 
	// MarshalJSON is called over and over again 
	contents, err := json.Marshal(map[string]interface{}{
		"IsNull": a.IsNull,
		"Accuracy": a.Accuracy,
		"City": a.City,
		"Country": a.Country,
		"CountryCode": a.CountryCode,
		"Latitude": a.Latitude,
		"Longitude": a.Longitude,
		"PostalCode": a.PostalCode,
		"State": a.State,
		"StateCode": a.StateCode,
		"Street": a.Street, 
	})

	if err != nil {
		return nil, err 
	}

	return contents, nil 
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (a Address) UnmarshalText(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		a.IsNull = true
		return nil
	}

	var err error
	err = json.Unmarshal(data, &a)
	if err != nil {
		return fmt.Errorf("UnmarshalText for NullableBool: %w", err)
	}

	return nil
}

func (a Address) String() string {
	contents, _ := json.MarshalIndent(a, "\t", "")
	return string(contents)
}
