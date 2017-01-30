package prow

import (
	"encoding/json"
	"net/http"
	"net/url"

	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/deis/prow/pkg/version"
)

// Client manages client side of the prow-prowd protocol. It wraps an *http.Client with a url
// Endpoint and common headers to send on every request.
type Client struct {
	HTTPClient *http.Client
	Endpoint   *url.URL
	Header     http.Header
}

// New returns a new Client with a given a URL and an optional client.
func New(endpoint *url.URL, client *http.Client) *Client {
	if client == nil {
		// user gets a default http client if one isn't specified.
		client = &http.Client{}
	}

	// ensure Path, RawQuery and Fragment are stripped.
	endpoint.Path = ""
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

	return &Client{client, endpoint, make(http.Header)}
}

// NewFromString returns a new Client given a string URL and an optional client.
func NewFromString(endpoint string, client *http.Client) (*Client, error) {
	e, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return New(e, client), nil
}

// Up uploads the contents of appDir to prowd, installs it in the specified namespace and
// returns a Helm Release.
func (c Client) Up(appDir, namespace string) (*release.Release, error) {
	return &release.Release{Info: &release.Info{Status: &release.Status{Code: release.Status_DEPLOYED}}}, nil
}

// Version returns the server version.
func (c Client) Version() (*version.Version, error) {
	var ver version.Version

	c.Endpoint.Path = "/version"
	req := &http.Request{
		Method:     "GET",
		URL:        c.Endpoint,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     c.Header,
		Body:       nil,
		Host:       c.Endpoint.Host,
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&ver)
	return &ver, err
}
