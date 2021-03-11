package requests_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/stretchr/testify/require"
)

type reqInput struct {
	builder requests.Builder
}

type reqOutput struct {
	description string
	requests    []*http.Request
	err         error
}

var reqSingularTest = map[*reqInput]reqOutput{
	{requests.Base.
		URL("https://google.com")}: {
		description: "basic GET request",
		requests: []*http.Request{{
			Method: http.MethodGet,
			URL:    requests.MustURL("https://google.com", nil),
			Header: http.Header{},
		}},
		err: nil,
	},
	{requests.Base.
		URL("https://google.com").
		Values(url.Values{"q": []string{"select foo from bar"}})}: {
		description: "basic GET request with querystring",
		requests: []*http.Request{{
			Method: http.MethodGet,
			URL:    requests.MustURL("https://google.com", &url.Values{"q": []string{"select foo from bar"}}),
			Header: http.Header{},
		}},
		err: nil,
	},
}

func compareRequest(t *testing.T, first, second *http.Request) error {
	if first.URL.String() != second.URL.String() {
		return fmt.Errorf("%s != %s", first.URL.String(), second.URL.String())
	}

	if first.Method != second.Method {
		return fmt.Errorf("%s != %s", first.Method, second.Method)
	}

	if first.Body != nil && second.Body == nil {
		return fmt.Errorf("second request is missing a body")
	}

	if first.Body != nil && second.Body != nil {
		b1, err := ioutil.ReadAll(first.Body)
		b2, err := ioutil.ReadAll(second.Body)

		if err != nil {
			return fmt.Errorf("failed to read body of request")
		}

		if bytes.Compare(b1, b2) != 0 {
			return fmt.Errorf("bodies are not equal")
		}
	}

	require.Equal(t, first.Header, second.Header, "header must be equal")
	require.Equal(t, first.URL.Query(), second.URL.Query(), "url.Values must be equal")

	return nil
}

func TestRequestBuilder(t *testing.T) {
	for in, out := range reqSingularTest {
		t.Run(out.description, func(t *testing.T) {
			req, err := in.builder.Request()
			require.Equal(t, out.err, err)
			require.Equal(t, 1, len(out.requests), "len(out.requests) must be equal to len(requests)")
			require.Nil(t, compareRequest(t, out.requests[0], req))
		})
	}
}
