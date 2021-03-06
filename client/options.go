package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beeekind/go-salesforce-sdk/internal/async"
	"github.com/beeekind/go-salesforce-sdk/requests"
	"github.com/dgrijalva/jwt-go"
)

// Option is a functional option used to configure the client object with
// a clearly documented and well defaulted approach: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
//
// Options are applied after salesforce client authentication so they have full access to method calls
// on the client in order to dynamically retrieve configuration information from the REST API itself
// if desired
type Option func(client *Client) error

// WithLoginURL sets the login url used to authenticate salesforce.Client
//
// Default: https://login.salesforce.com/services/oauth2/token
func WithLoginURL(loginURL string) Option {
	return func(client *Client) error {
		client.loginURL = loginURL
		return nil
	}
}

// LoginResponse ...
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	ID          string `json:"id"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	Signature   string `json:"signature"`
}

// LoginError ...
type LoginError struct {
	ErrorMessage     string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *LoginError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorMessage, e.ErrorDescription)
}

// WithLoginFailover ... 
func WithLoginFailover(options ...Option) Option {
	return func(client *Client) error {
		var err error 
		for _, opt := range options {
			err = opt(client)
			if err == nil {
				return nil 
			}
		}
		return err 
	}
}

// WithPasswordBearer ...
func WithPasswordBearer(clientID, clientSecret, username, password, securityToken string) Option {
	return func(client *Client) error {
		var loginResponse *LoginResponse
		_, err := requests.
			URL(client.loginURL).
			Method(http.MethodPost).
			Header("Content-Type", "application/x-www-form-urlencoded").
			Values(url.Values{
				"grant_type":    []string{"password"},
				"client_id":     []string{clientID},
				"client_secret": []string{clientSecret},
				"username":      []string{username},
				"password":      []string{password},
			}).
			JSON(&loginResponse)

		if err != nil {
			return err
		}

		return WithLoginResponse(loginResponse)(client)
	}
}

// WithJWTBearer ...
func WithJWTBearer(clientID, clientUsername, privateKeyPath string) Option {
	return func(client *Client) error {
		contents, err := ioutil.ReadFile(privateKeyPath)
		if err != nil {
			return err
		}

		signature, err := jwt.ParseRSAPrivateKeyFromPEM(contents)
		if err != nil {
			return err
		}

		claims := jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Audience:  client.loginURL,
			Issuer:    clientID,
			Subject:   clientUsername,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenString, err := token.SignedString(signature)
		if err != nil {
			return err
		}

		var loginResponse *LoginResponse
		_, err = requests.
			URL(client.loginURL).
			Method(http.MethodPost).
			Header("Content-Type", "application/x-www-form-urlencoded").
			Values(url.Values{
				"grant_type": []string{"urn:ietf:params:oauth:grant-type:jwt-bearer"},
				"assertion":  []string{tokenString},
				"client_id":  []string{clientID},
			}).
			JSON(&loginResponse)

		if err != nil {
			return err
		}

		return WithLoginResponse(loginResponse)(client)
	}
}

// WithLoginResponse derives needed URL components used in all subsequent requests to
// Salesforce including your salesforce instanceURL, authorization bearer token, and
// url path prefix ("/services/data")
func WithLoginResponse(loginResponse *LoginResponse) Option {
	return func(client *Client) error {
		if err := WithInstanceURL(loginResponse.InstanceURL)(client); err != nil {
			return fmt.Errorf("WithLoginResponse(): %s, %w", loginResponse.InstanceURL, err)
		}

		if err := WithHTTPClient(NewHTTPClient(
			TransportWithHeader("Authorization", "Bearer "+loginResponse.AccessToken),
			TransportWithHeader("Accept-Encoding", "gzip"),
		))(client); err != nil {
			return fmt.Errorf("WithLoginResponse(): %w", err)
		}

		versions, err := client.APIVersions()
		if err != nil {
			return fmt.Errorf("WithLoginResponse(): %w", err)
		}

		if len(versions) == 0 {
			return errors.New("WithLoginResponse(): len(versions) == 0")
		}

		latestVersion := versions[len(versions)-1]
		parts := strings.Split(latestVersion.URL, "/")

		// expected output is like []string{"", "services", "data", "v50.0"}
		if len(parts) != 4 {
			return errors.New("WithLoginResponse(): len(parts) != 4")
		}

		if err := WithVersion(latestVersion.Version)(client); err != nil {
			return fmt.Errorf("WithLoginResponse(): %s: %w", latestVersion.Version, err)
		}

		return nil
	}
}

// WithInstanceURL sets the instance url representing your organizations unique hostname for
// accessing the salesforce REST API
//
// a developer organization may look like https://{{ORGANIZATION_NAME}}-dev-ed.my.salesforce.com/
// an enterprise organization may look like https://na150.salesforce.com/
func WithInstanceURL(instanceURL string) Option {
	return func(client *Client) error {
		if instanceURL == "" {
			return errors.New("WithInstanceURL(): instanceURL == \"\"")
		}
		client.instanceURL = instanceURL
		return nil
	}
}

// WithHTTPClient sets the client used to access the Salesforce REST API. This client should be configured
// similar to httppool/authhttp.go where each request includes the proper authorization headers. See
// WithLoginBearer for an example of how to property configure the http.Client through authhttp
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) error {
		if httpClient == nil {
			return errors.New("WithHTTPClient: httpClient must not be nil")
		}
		client.client = httpClient
		return nil
	}
}

// WithUsage is the percentage of your organizations daily API requests that this library
// will use before returning errors. This is calculated by reading the "Sforce-Limit-Info"
// header returned by some types of requests. Until the header is found in a made request
// this limit will fall back to the WithDailyAPIMax property.
//
// Default 0.70 representing 70%
func WithUsage(apiUsagePercentage float64) Option {
	return func(client *Client) error {
		if apiUsagePercentage == 0 {
			return errors.New("WithUsage(): apiUsagePercentage cannot be 0")
		}
		client.apiUsageLimit = apiUsagePercentage
		return nil
	}
}

// WithDailyAPIMax is the maximum number of requests allowed in a 24 hour period as seen
// in the company information section of your salesforce setup page. While the Salesforce client
// will automatically determine your daily api limit based on the "Sforce-Limit-Info" header it
// is useful to have a default value for calculating the apiUsageLimit threshold when the client
// is first initialized or if the header cannot be found.
//
// 15,000 is the default limit when you create a fresh Salesforce developer account as of 12/2020.
// If your limit is reached many Salesforce services will begin returning errors which could disrupt
// your organizations use of the Salesforce platform entirely.
//
// Salesforce will grant additional requests in emergency situations if you contact your
// salesforce representative though generally by the time they respond it will be too late.
//
// Default: 15000
func WithDailyAPIMax(maxRequests24hr int64) Option {
	return func(client *Client) error {
		if maxRequests24hr == 0 {
			return errors.New("WithDailyAPIMax(): maxRequests24hr cannot be 0")
		}
		client.dailyAPILimit = maxRequests24hr
		return nil
	}
}

// WithPathPrefix is the prefix used to form a a fully qualified URL for retrieving data from
// the Salesforce REST API. By convention this will be /services/data but in the case that future
// API versions choose a different format we're leaving this as a dynamically configurable option.
//
// Note that we trim any leading or trailing "/" characters since we later join it with other
// url segments when preparing the full url.
//
// Default: "services/data"
func WithPathPrefix(urlTemplatePrefix string) Option {
	return func(client *Client) error {
		urlTemplatePrefix = strings.TrimPrefix(urlTemplatePrefix, "/")
		urlTemplatePrefix = strings.TrimSuffix(urlTemplatePrefix, "/")
		client.apiPathPrefix = urlTemplatePrefix
		return nil
	}
}

// WithVersion sets the api version for subsequent requests made to the Salesforce REST API. By default
// it will use the latest API version which is v50.0 as of 12/2020.
//
// If you wish to make requests across multiple versions create multiple
// instances of a salesforce.Client, one for each desired version.
//
// Salesforce keeps backwards compatibility better then many other services,
// though you should read the individual SLA for any service whose backwards compatibility
// is vital to your organization.
//
// version is formatted as: "v50.0"
//
// Default: v50.0
func WithVersion(apiVersion string) Option {
	return func(client *Client) error {
		if apiVersion == "" {
			return errors.New("WithVersion(): apiVersion cannot be \"\"")
		}

		if strings.HasPrefix(apiVersion, "v") {
			return errors.New("WithVersion(): apiVersion should not have v prefix, the v prefix is applied in this method")
		}

		if !strings.HasSuffix(apiVersion, ".0") {
			return errors.New("WithVersion(): apiVersion should be a string formatted as a float with a single decimal precision i.e. 50.0")
		}

		client.apiVersion = fmt.Sprintf("v%s", apiVersion)
		return nil
	}
}

// WithPool ...
func WithPool(pool *async.Pool) Option {
	return func(client *Client) error {
		client.pool = pool
		return nil
	}
}

// WithLimiter ...
func WithLimiter(limiter Limiter) Option {
	return func(client *Client) error {
		client.limiter = limiter
		return nil
	}
}
