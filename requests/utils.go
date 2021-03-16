package requests

// http.go contains project specific utilities that improve on the standard libraries net/http library

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ErrUnmarshalEmpty ...
var ErrUnmarshalEmpty = errors.New("could not unmarshal an empty buffer")

// RequestError contains additional metadata about the http error
type RequestError struct {
	Code     int
	Contents []byte
	Err      error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("http response (%v) returned: %s", e.Code, e.Contents)
}

// Is ...
func (e *RequestError) Is(tgt error) bool {
	_, ok := tgt.(*RequestError)
	return ok
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

// Unmarshal ...
func (e *RequestError) Unmarshal(dst interface{}) error {
	return json.Unmarshal([]byte(e.Contents), dst)
}

// ReadAndCloseResponse is a safety function that ensures we do not leak system resources by closing
// the http.Response Body after it is read
//
// The semantics of body.Close() are a little confusing in Golang and because of the quantity of http requests we
// make in our applications it is easy to run out of system resources such as file descriptors
func ReadAndCloseResponse(r *http.Response) ([]byte, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("ReadAndCloseResponse(): %w", err)
		}

		r.Body = reader
	}

	defer r.Body.Close()

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAndCloseResponse(): %w", err)
	}

	if r.StatusCode >= 299 {
		return contents, &RequestError{r.StatusCode, contents, fmt.Errorf("ReadAndCloseResponse(): StatusCode: %v", r.StatusCode)}
	}

	return contents, nil
}

// Unmarshal reads and closes an http response body
// it returns the raw bytes of the body and any error which occurred
//
// this utility function reduces the boilerplate code necessary for unmarshaling
// http.Response bodies
func Unmarshal(r *http.Response, dst interface{}) ([]byte, error) {
	contents, err := ReadAndCloseResponse(r)

	var e *RequestError
	if err != nil && !errors.Is(err, e) {
		return contents, fmt.Errorf("requests.Unmarshal(): %w", err)
	}

	if err != nil {
		return nil, err
	}

	if len(contents) == 0 {
		return nil, ErrUnmarshalEmpty
	}

	if err := json.Unmarshal(contents, dst); err != nil {
		return contents, fmt.Errorf("requests.Unmarshal(): %w", err)
	}

	return contents, nil
}

// ExtractUsageHeader extracts two integer value from the Sforce-Limit-Info header which contains
// a formatted string containing that information
//
// Because we are extracting integer data from a complex string, this process is prone to error if
// an unexpected value is found
func ExtractUsageHeader(response *http.Response) (used int64, total int64, err error) {
	limits := response.Header.Get("Sforce-Limit-Info")
	if limits == "" {
		return 0, 0, errors.New("failed to ExtractUsageHeader(): Sforce-Limit-Info header does not exist")
	}

	usedSlashRemaining := strings.Split(limits, "=")

	if len(usedSlashRemaining) != 2 {
		return 0, 0, errors.New("failed to ExtractUsageHeader(): invalid result of strings.Split(=)")
	}

	usedRemaining := strings.Split(usedSlashRemaining[1], "/")
	u, err := strconv.Atoi(usedRemaining[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to ExtractUsageHeader():%w", err)
	}

	t, err := strconv.Atoi(usedRemaining[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to ExtractUsageHeader():%w", err)
	}

	return int64(u), int64(t), nil
}

// ComputeSubsequentRecordURLs identifies the cursor used for a query
// with multiple results pages and computes all subsequent page urls from
// the first two requests
//
// i.e.
// query /query?q=select Name from Lead
// => returns firstNextRecordsURL: "some-url-1000"
// query firstNextRecordsURL
// => returns secondNextRecordsURL "some-url-2000" <-- calculate the increment
// computeSubsequentRecordURLs(firstNextRecordsURL, secondNextRecordsURL)
//
// this allows us to then query asynchronously the remaining records asynchronously
// rather then walking them one at a time
//
// NOTE: this behavior is dependant on salesforce pagination being predetermined after
// the first two requests. as far as I know, this implementation is undocumented and could
// change at any time. though it is VERY convenient and offers a massive performance improvement.
func ComputeSubsequentRecordURLs(instanceURL string, firstNextRecordsURL string, secondNextRecordsURL string, totalRecords int) (URLs []string, err error) {
	URLs = []string{instanceURL + secondNextRecordsURL}
	uri, firstInterval, err := computeInterval(firstNextRecordsURL)
	if err != nil {
		return nil, fmt.Errorf("ComputeSubsequentRecordURLs(1): %w", err)
	}

	_, secondInterval, err := computeInterval(secondNextRecordsURL)
	if err != nil {
		return nil, fmt.Errorf("ComputeSubsequentRecordURLs(2): %w", err)
	}

	querySize := secondInterval - firstInterval
	nextInterval := secondInterval + querySize

	for nextInterval <= totalRecords {
		strInterval := strconv.Itoa(nextInterval)
		url := fmt.Sprintf("%s%s-%s", instanceURL, uri, strInterval)
		println(5, url)
		URLs = append(URLs, url)
		nextInterval = nextInterval + querySize
	}

	return URLs, nil
}

// ComputeInterval ...
func computeInterval(URI string) (uri string, interval int, err error) {
	elems := strings.Split(URI, "/")
	if len(elems) != 6 {
		return "", 0, fmt.Errorf("computeInterval(1): computeInterval len(elems) %v != %v: %s", len(elems), 6, URI)
	}

	parts := strings.Split(elems[len(elems)-1], "-")
	if len(parts) != 2 {
		return "", 0, errors.New("computeInterval(2): could not split into 2 pieces by -:" + elems[len(elems)-1])
	}

	uri = strings.Join(elems[:len(elems)-1], "/") + "/" + parts[0]
	interval, err = strconv.Atoi(parts[1])
	if err != nil {
		return uri, 0, fmt.Errorf("computeInterval(3): %w", err)
	}

	return uri, interval, err
}

// MustURL ...
func MustURL(str string, values *url.Values) *url.URL {
	url, err := url.Parse(str)
	if err != nil {
		panic(err)
	}

	if values != nil {
		url.RawQuery = values.Encode()
	}

	return url
}
