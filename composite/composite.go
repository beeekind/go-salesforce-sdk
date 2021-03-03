package composite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/b3ntly/salesforce/metadata"
	"github.com/b3ntly/salesforce/requests"
	"github.com/lann/builder"
)

// package composite uses the builder pattern to compose
// composite api requests for the Salesforce REST API
//
// Documentation:
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_composite.htm

// Builder ..
type Builder builder.Builder

type client interface {
	Do(req *http.Request) (*http.Response, error)
	URL(partial string) string
}

// Base ...
var Base = Builder(builder.EmptyBuilder)

func init() {
	builder.Register(Builder{}, Request{})
}

// SQLizer defines a component which can be converted to SQL.
type SQLizer interface {
	ToSQL() (soql string, err error)
}

// Request ...
type Request struct {
	Client             client       `json:"-"`
	AllOrNone          bool         `json:"allOrNone"`
	CollateSubrequests bool         `json:"collateSubrequests"`
	CompositeRequest   []Subrequest `json:"compositeRequest"`
}

// Response ...
type Response struct {
	Items []*ResponseItem `json:"compositeResponse"`
}

// Err returns an error if one or more Response.Items contains a HTTPStatusCode > 299
func (r Response) Err() error {
	if len(r.Items) == 0 {
		return errors.New("composite.Response contained no response items")
	}

	for _, item := range r.Items {
		if item.HTTPStatusCode > 299 {
			var buff strings.Builder
			for _, o := range item.Outputs {
				msg := fmt.Sprintf("%v: %s: %s", item.HTTPStatusCode, o.ErrorCode, o.Message)
				buff.WriteString(msg)
			}
			return errors.New(buff.String())
		}
	}
	return nil
}

// ResponseItem https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/responses_composite.htm
type ResponseItem struct {
	Outputs        []*Output
	HTTPHeaders    map[string]interface{} `json:"httpHeaders,omitempty"`
	HTTPStatusCode int                    `json:"httpStatusCode"`
	ReferenceID    string                 `json:"referenceId"`
	Body           map[string]interface{} `json:"body"`
}

// Output ...
type Output struct {
	ID        string `json:"id,omitempty"`
	Success   bool   `json:"success,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`
	Message   string `json:"message,omitempty"`
}

// UnmarshalJSON ...
func (item *ResponseItem) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for k, v := range raw {
		if v == nil {
			continue
		}

		var ok bool
		var err error
		switch k {
		case "httpHeaders":
			if item.HTTPHeaders, ok = v.(map[string]interface{}); !ok {
				return errors.New("composite.ResponseItem custom UnmarshalJSON failed to caste property " + k)
			}
		case "httpStatusCode":
			if elem, ok := v.(float64); !ok {
				return errors.New("composite.ResponseItem custom UnmarshalJSON failed to caste property " + k)
			} else {
				item.HTTPStatusCode = int(elem)
			}
		case "referenceId":
			if item.ReferenceID, ok = v.(string); !ok {
				return errors.New("composite.ResponseItem custom UnmarshalJSON failed to caste property " + k)
			}
		case "body":
			switch v.(type) {
			case map[string]interface{}:
				payload, _ := json.Marshal(v)
				var output Output
				if err := json.Unmarshal(payload, &output); err != nil {
					return err
				}
				item.Outputs = []*Output{&output}
			case []interface{}:
				payload, _ := json.Marshal(v)
				if err := json.Unmarshal(payload, &item.Outputs); err != nil {
					return err
				}
			default:
				return errors.New("composite.ResponseItem custom UnmarshalJSON failed to caste property " + k)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Error ...
type Error struct {
	StatusCode string   `json:"statusCode,omitempty"`
	ErrorCode  string   `json:"errorCode,omitempty"`
	Message    string   `json:"message,omitempty"`
	Fields     []string `json:"fields,omitempty"`
}

// Errors ...
type Errors []Error

func (e Errors) Error() string {
	var buff strings.Builder
	for _, err := range e {
		buff.WriteString(fmt.Sprintf("%s: %v\n", err.Message, err.Fields))
	}
	return buff.String()
}

// Attributes are metadata properties used in certain APIs such as the composite API
type Attributes struct {
	Type        string `json:"type"`
	ReferenceID string `json:"referenceId"`
}

// Subrequest ...
type Subrequest struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	ReferenceID string                 `json:"referenceId"`
	Body        map[string]interface{} `json:"body,omitempty"`
	HTTPHeaders map[string]string      `json:"httpHeaders,omitempty"`
}

// Request marshals the given Builder into an http.Request object
func (b Builder) Request() (*http.Request, error) {
	data := builder.GetStruct(b).(Request)
	return data.request()
}

// request ...
func (r Request) request() (*http.Request, error) {
	if r.Client == nil {
		return nil, fmt.Errorf("CompositeBuilder must have a non-nil client to call Response()")
	}

	// normalize urls
	for i := 0; i < len(r.CompositeRequest); i++ {
		// https://placeholder-dev-ed.my.salesforce.com/services/data/v51.0/composite/sobjects/Lead
		// =>
		// /services/data/v51.0/composite/sobjects/Lead
		parts := strings.Split(r.Client.URL(r.CompositeRequest[i].URL), "/")
		r.CompositeRequest[i].URL = "/" + strings.Join(parts[3:], "/")
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshaling composite.Request: %w", err)
	}

	buff := bytes.NewBuffer(payload)
	req, err := http.NewRequest(http.MethodPost, r.Client.URL(metadata.CompositeEndpoint), buff)
	if err != nil {
		return nil, fmt.Errorf("converting composite.Request to http.Request: %w", err)
	}

	//
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

// Response ...
func (b Builder) Response() (*http.Response, error) {
	data := builder.GetStruct(b).(Request)

	if data.Client == nil {
		return nil, fmt.Errorf("CompositeBuilder must have a non-nil client to call Response()")
	}

	req, err := b.Request()
	if err != nil {
		return nil, err
	}

	return data.Client.Do(req)
}

// JSON ...
func (b Builder) JSON(dst interface{}) (body []byte, err error) {
	response, err := b.Response()
	if err != nil {
		return nil, err
	}

	return requests.Unmarshal(response, dst)
}

// Send ...
func (b Builder) Send() (*Response, error) {
	var response Response
	if _, err := b.JSON(&response); err != nil {
		return nil, err
	}

	return &response, response.Err()
}

// Client ...
func Client(client client) Builder {
	return Base.Client(client)
}

// Client ...
func (b Builder) Client(client client) Builder {
	return builder.Set(b, "Client", client).(Builder)
}

// AllOrNone ...
func AllOrNone(allOrNone bool) Builder {
	return Base.AllOrNone(allOrNone)
}

// AllOrNone ...
func (b Builder) AllOrNone(allOrNone bool) Builder {
	return builder.Set(b, "AllOrNone", allOrNone).(Builder)
}

// CollateSubrequests ...
func CollateSubrequests(collateSubrequests bool) Builder {
	return Base.CollateSubrequests(collateSubrequests)
}

// CollateSubrequests ...
func (b Builder) CollateSubrequests(collateSubrequests bool) Builder {
	return builder.Set(b, "CollateSubrequests", collateSubrequests).(Builder)
}

// Add creates a new Subrequest instance from the given parameters and appends the instance
// to builder.CompositeRequest ([]Subrequest)
func (b Builder) Add(method string, URL string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	subrequest := Subrequest{
		Method:      method,
		URL:         URL,
		ReferenceID: referenceID,
		Body:        body,
		HTTPHeaders: headers,
	}
	return builder.Append(b, "CompositeRequest", subrequest).(Builder)
}

// Get ...
func Get(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return Base.Add(http.MethodGet, fmt.Sprintf("%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Get ...
func (b Builder) Get(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return b.Add(
		http.MethodGet,
		fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName),
		referenceID,
		headers,
		body,
	)
}

// Post ...
func Post(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return Base.Add(http.MethodPost, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Post ...
func (b Builder) Post(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return b.Add(http.MethodPost, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Patch ...
func Patch(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return Base.Add(http.MethodPatch, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Patch ...
func (b Builder) Patch(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return b.Add(http.MethodPatch, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Delete ...
func Delete(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return Base.Add(http.MethodDelete, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}

// Delete ...
func (b Builder) Delete(objectName string, referenceID string, headers map[string]string, body map[string]interface{}) Builder {
	return b.Add(http.MethodDelete, fmt.Sprintf("/%s/%s/", metadata.SobjectsEndpoint, objectName), referenceID, headers, body)
}
