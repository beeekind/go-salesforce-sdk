// Package client wraps http.Client and handles request authentication and other session state
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/beeekind/go-salesforce-sdk/soql"
	"github.com/beeekind/go-salesforce-sdk/types"
)

const (
	EnvSalesForceSDKInstanceURL    = "SALESFORCE_SDK_INSTANCE_URL"
	EnvSalesForceSDKLoginURL       = "SALESFORCE_SDK_LOGIN_URL"
	EnvSalesForceSDKPathPrefix     = "SALESFORCE_SDK_PATH_PREFIX"
	EnvSalesForceSDKClientID       = "SALESFORCE_SDK_CLIENT_ID"
	EnvSalesForceSDKClientSecret   = "SALESFORCE_SDK_CLIENT_SECRET"
	EnvSalesForceSDKUsername       = "SALESFORCE_SDK_USERNAME"
	EnvSalesForceSDKPassword       = "SALESFORCE_SDK_PASSWORD"
	EnvSalesForceSDKSecurityToken  = "SALESFORCE_SDK_SECURITY_TOKEN"
	EnvSalesForceJWTPrivateKeyPath = "SALESFORCE_SDK_JWT_PRIVATE_KEY_PATH"

	DefaultLoginURL = "https://login.salesforce.com/services/oauth2/token"

	DefaultPathPrefix = "services/data"
)

// Client ...
type Client struct {
	// mu protects access to reference fields in this struct like client, pool, and limiter
	limiter       Limiter
	client        *http.Client
	loginURL      string
	instanceURL   string
	apiPathPrefix string
	apiVersion    string
	apiUsageLimit float64
	dailyAPILimit int64
	usedAPILast24 int64
}

// Limiter defines behavior for ratelimiting outgoing http requests to the Salesforce API
type Limiter interface {
	Allow(key string) (nextAllowed time.Duration, err error)
}

// APIVersion ...
type APIVersion struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

var defaultOptions = []Option{
	WithFirstSuccess(
		WithLoginURLFromEnv(),
		WithLoginURL(DefaultLoginURL),
	),
	WithFirstSuccess(
		WithPathPrefixFromEnv(),
		WithPathPrefix(DefaultPathPrefix),
	),
	WithDailyAPIMax(15000),
	WithUsage(0.40),
}

// Must calls New(options...) and panics if an error occurs
func Must(options ...Option) *Client {
	client, err := New(options...)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new instance of client.Client . The instance may be customized
// by passing in client.Option types.
func New(options ...Option) (*Client, error) {
	client := &Client{}

	// apply default options first
	for _, opt := range defaultOptions {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	// apply custom options last
	for _, opt := range options {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	// if no authentication option has been initiated default to password based authentication
	// using environmental variables
	if client.client == nil {
		err := WithPasswordBearer(
			os.Getenv("SALESFORCE_SDK_CLIENT_ID"),
			os.Getenv("SALESFORCE_SDK_CLIENT_SECRET"),
			os.Getenv("SALESFORCE_SDK_USERNAME"),
			os.Getenv("SALESFORCE_SDK_PASSWORD"),
			os.Getenv("SALESFORCE_SDK_SECURITY_TOKEN"),
		)(client)

		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// IsWithinAPIUsageLimit ...
func (c *Client) IsWithinAPIUsageLimit() error {
	used := atomic.LoadInt64(&c.usedAPILast24)
	total := atomic.LoadInt64(&c.dailyAPILimit)
	usage := float64(used) / float64(total)

	if total == 0 {
		return nil
	}

	if usage > c.apiUsageLimit {
		return fmt.Errorf("salesforce API usage rate exceeded %d/%d (limit: %f%% actual: %f%%)", used, total, c.apiUsageLimit, usage*float64(100))
	}

	return nil
}

// Do proxies call to http.Client.Do with the following extended behavior:
// * Requests are only made if IsWithinAPIUsageLimit does not return an error
// * Any configured Limiter allows a request
// * Responses are parsed in order to update client.UsedAPILast24
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if err := c.IsWithinAPIUsageLimit(); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// derived from the response header, keep our current api usage up to date
	// IMPORTANT: some requests will not return a valid usage header included attempts to
	// access an endpoint which has not been enabled for the salesforce account you're using
	apiRequestsUsed, apiRequestsTotal, err := requests.ExtractUsageHeader(resp)
	if err != nil {
		return resp, nil
	}

	atomic.StoreInt64(&c.usedAPILast24, apiRequestsUsed)
	atomic.StoreInt64(&c.dailyAPILimit, apiRequestsTotal)
	return resp, nil
}

// APIVersions returns a list of all available Salesforce versions. Generally
// you'll want to use the latest API version as Salesforce puts a lot of effort into
// backwards compatibility.
//
// This method is called by WithLoginResponse() in order to select
// the latest API version by default.
func (c *Client) APIVersions() (versions []*APIVersion, err error) {
	_, err = requests.
		Sender(c).
		URL(fmt.Sprintf("%s/%s", c.instanceURL, c.apiPathPrefix)).
		JSON(&versions)

	if err != nil {
		return nil, err
	}

	return versions, nil
}

// QueryMore executes a soql query on the query endpoint and returns all paginated
// resources. By analyzing the first and second serialized requests
// we can pre-compute all subsequent paginated resources and concurrently process remaining
// work.
//
// This concurrent approach provides a massive performance increase when
// querying many records.
func (c *Client) QueryMore(builder soql.Builder, dst interface{}, includeSoftDelete bool) (err error) {
	// 1) make the initial query and retrieve metadata on the total size
	var firstResponse types.QueryResponse
	path := "query"
	if includeSoftDelete {
		path = "queryAll"
	}

	_, err = requests.Sender(c).URL(path).SQLizer(builder).JSON(&firstResponse)
	if err != nil {
		return err
	}

	if firstResponse.Done {
		return json.Unmarshal(firstResponse.Records, dst)
	}

	// 2) make a second query
	var secondResponse types.QueryResponse
	_, err = requests.Sender(c).URL(firstResponse.NextRecordsURL).JSON(&secondResponse)
	if err != nil {
		return err
	}

	if secondResponse.Done {
		return json.Unmarshal(
			requests.MergeJSONArrays(firstResponse.Records, secondResponse.Records),
			dst,
		)
	}

	initialPayloads := [][]byte{firstResponse.Records, secondResponse.Records}

	// 3) use the first and second queries to compute all paginated resources
	URLs, err := requests.ComputeSubsequentRecordURLs(c.instanceURL, firstResponse.NextRecordsURL, secondResponse.NextRecordsURL, firstResponse.TotalSize)
	if err != nil {
		return err
	}

	// 4) execute all subsequent paginated queries
	payloads, err := c.querySubsequentURLs(URLs...)
	if err != nil {
		return err
	}

	// 5) build them into a single []byte which can be unmarshalled
	results := requests.MergeJSONArrays(append(initialPayloads, payloads...)...)

	// 6) unmarshal them
	return json.Unmarshal(results, dst)
}

type result struct {
	Body []byte
	Err  error
}

func (c *Client) querySubsequentURLs(paginatedURLs ...string) (payloads [][]byte, err error) {
	numWorkers := 100
	if numWorkers > len(paginatedURLs) {
		numWorkers = len(paginatedURLs)
	}

	input := make(chan string, len(paginatedURLs))
	output := make(chan *result, len(paginatedURLs))

	for i := 0; i < len(paginatedURLs); i++ {
		item := paginatedURLs[i]
		input <- item
	}
	close(input)

	var wg sync.WaitGroup
	wg.Add(len(paginatedURLs))

	for j := 0; j < numWorkers; j++ {
		go func(wg *sync.WaitGroup, client *Client, input chan string, output chan *result) {
			for url := range input {
				contents, err := requests.Sender(client).URL(url).JSON(nil)
				output <- &result{contents, err}
				wg.Done()
			}

		}(&wg, c, input, output)
	}

	wg.Wait()
	close(output)

	for result := range output {
		if result.Err != nil {
			return nil, err
		}

		payloads = append(payloads, result.Body)
	}

	return payloads, nil
}

// URL parses a url segment into a fully qualified Salesforce API request using client.instanceURL,
// client.apiPathPrefix, and client.apiVersion
//
// Any fully qualified url - as indicated by the presence of the string https - is returned
// unmodified.
func (c *Client) URL(path string) string {
	if strings.Contains(path, "https") {
		return path
	}

	//
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	instanceURL := c.instanceURL
	prefix := c.apiPathPrefix
	version := c.apiVersion
	URL := strings.Join([]string{instanceURL, prefix, version, path}, "/")
	return URL
}
