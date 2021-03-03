package client

// authhttp.go creates http.Client configured to reuse authorization headers alongside better defaults

import (
	"net"
	"net/http"
	"runtime"
	"time"
)

// RoundTripFunc ...
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip ...
func (fn RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// TransportOption represents a transport-level option for an http.RoundTripper.
type TransportOption func(http.RoundTripper) http.RoundTripper

// TransportWithBasicAuth calls http.Request.SetBasicAuth with the given values for each request
func TransportWithBasicAuth(clientID, secret string) TransportOption {
	return func(rt http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.SetBasicAuth(clientID, secret)
			return rt.RoundTrip(req)
		})
	}
}

// TransportWithHeader appends the following key/value header to each subsequent request
func TransportWithHeader(key, value string) TransportOption {
	return func(rt http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set(key, value)
			return rt.RoundTrip(req)
		})
	}
}

// https://blog.cloudflare.com/exposing-go-on-the-internet/
func newBaseRoundTripper() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

// NewHTTPClient constructs a new HTTP client with the specified transport-level options.
func NewHTTPClient(opts ...TransportOption) *http.Client {
	rt := newBaseRoundTripper()
	for _, opt := range opts {
		rt = opt(rt)
	}
	return &http.Client{Transport: rt}
}
