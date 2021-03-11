package requests

// requests.go provides a fluent api for generating http requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/lann/builder"
)

// Base ...
var Base = Builder(builder.EmptyBuilder).Sender(DefaultSender)

// Sender ...
func Sender(sender sender) Builder {
	return Base.Sender(sender)
}

// URL ...
func URL(url string) Builder {
	return Base.URL(url)
}

// JSON ...
var JSON = Base.Header("Content-Type", "application/json")

// JSONURL ...
func JSONURL(url string) Builder {
	return JSON.URL(url)
}

// init is necessary for preparing data structures used by builder.Builder
func init() {
	builder.Register(Builder{}, requestData{})
}

type sender interface {
	Do(req *http.Request) (*http.Response, error)
	QueryMore(builder soql.Builder, dst interface{}, includeSoftDelete bool) error
	URL(string) string
}

var DefaultSender = &defaultSender{}

type defaultSender struct {
	http.Client
}

func (s *defaultSender) QueryMore(builder soql.Builder, dst interface{}, includeSoftDelete bool) error {
	return nil
}

func (s *defaultSender) URL(str string) string {
	return str
}

// Builder ...
type Builder builder.Builder

type requestData struct {
	// data structures for singular requests
	URL    string
	Method string
	Body   io.Reader
	Ctx    context.Context
	//
	Marshal interface{}
	Values  url.Values
	Header  http.Header
	SQLizer sqlizer
	//
	Sender sender
}

type sqlizer interface {
	ToSQL() (string, error)
}

// Request composes an http.Request from its base components:
// * a url string [REQUIRED]
// * an http method string [OPTIONAL, DEFAULT = http.MethodGet]
// * url parameters aka url.Values [OPTIONAL]
// * a request body [OPTIONAL]
// * http headers [OPTIONAL]
func (b Builder) Request() (*http.Request, error) {
	data := builder.GetStruct(b).(requestData)

	if data.Method == "" {
		data.Method = http.MethodGet
	}

	if data.Marshal != nil {
		contents, err := json.Marshal(data.Marshal)
		if err != nil {
			return nil, err
		}

		println(0, string(contents))
		data.Body = bytes.NewBuffer(contents)
	}

	if data.Sender != nil {
		data.URL = data.Sender.URL(data.URL)
	}

	// method, url, and body
	req, err := http.NewRequest(data.Method, data.URL, data.Body)
	if err != nil {
		return nil, fmt.Errorf("RequestBuilder could not make a request: %w", err)
	}

	// headers
	for key, values := range data.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// override values["q"] if sqlizer != nil
	if data.SQLizer != nil {
		sql, err := data.SQLizer.ToSQL()
		if err != nil {
			return nil, err
		}
		if data.Values == nil {
			data.Values = url.Values{}
		}

		data.Values.Add("q", sql)
	}

	// values
	if len(data.Values) > 0 {
		req.URL.RawQuery = data.Values.Encode()
	}

	if data.Ctx != nil {
		req = req.WithContext(data.Ctx)
	}

	return req, nil
}

// Response ...
func (b Builder) Response() (*http.Response, error) {
	data := builder.GetStruct(b).(requestData)

	if data.Sender == nil {
		return nil, fmt.Errorf("RequestBuilder must have a non nil sender to produce a http.Response")
	}

	req, err := b.Request()
	if err != nil {
		return nil, err
	}

	return data.Sender.Do(req)
}

// JSON ...
func (b Builder) JSON(dst interface{}) (body []byte, err error) {
	response, err := b.Response()
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, fmt.Errorf("for some reason the response is nil")
	}

	return Unmarshal(response, dst)
}

// QueryMore ...
func (b Builder) QueryMore(selectBuilder soql.Builder, dst interface{}, includeSoftDelete bool) error {
	data := builder.GetStruct(b).(requestData)
	if data.Sender == nil {
		return fmt.Errorf("requests.Builder must have a non nil sender to complete this action")
	}

	return data.Sender.QueryMore(selectBuilder, dst, includeSoftDelete)
}

// Sender ...
func (b Builder) Sender(sender sender) Builder {
	return builder.Set(b, "Sender", sender).(Builder)
}

// URL ...
func (b Builder) URL(str string) Builder {
	return builder.Set(b, "URL", str).(Builder)
}

// Method ...
func (b Builder) Method(method string) Builder {
	return builder.Set(b, "Method", method).(Builder)
}

// Context ...
func (b Builder) Context(ctx context.Context) Builder {
	return builder.Set(b, "Ctx", ctx).(Builder)
}

// Values ...
func (b Builder) Values(values url.Values) Builder {
	return builder.Set(b, "Values", values).(Builder)
}

// Param ...
func (b Builder) Param(key, value string) Builder {
	elem, exists := builder.Get(b, "Values")
	if !exists {
		return b.Values(url.Values{key: []string{value}})
	}

	values, ok := elem.(url.Values)
	if !ok {
		return b
	}

	values.Add(key, value)
	return b.Values(values)
}

// SQLizer ...
func (b Builder) SQLizer(sqlizer sqlizer) Builder {
	return builder.Set(b, "SQLizer", sqlizer).(Builder)
}

// Body ...
func (b Builder) Body(body io.Reader) Builder {
	return builder.Set(b, "Body", body).(Builder)
}

// Marshal ...
func (b Builder) Marshal(item interface{}) Builder {
	return builder.Set(b, "Marshal", item).(Builder)
}

// Headers ...
func (b Builder) Headers(header http.Header) Builder {
	return builder.Set(b, "Header", header).(Builder)
}

// Header appends a single header to any existing headers
func (b Builder) Header(key string, values ...string) Builder {
	elem, exists := builder.Get(b, "Header")
	if !exists {
		return b.Headers(http.Header{key: values})
	}

	headers, ok := elem.(http.Header)
	if !ok {
		return b
	}

	for i := 0; i < len(values); i++ {
		headers.Add(key, values[i])
	}

	return b.Headers(headers)
}
