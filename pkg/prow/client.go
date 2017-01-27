package prow

import (
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
	Header http.Header
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
	return &release.Release{}, nil
}

// Version returns the server version.
func (c Client) Version() (*version.Version, error) {
	return version.New(), nil
}
