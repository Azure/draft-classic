package prow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/deis/prow/pkg/version"
)

// closingBuffer is a bytes.Buffer that implements .Close() so we can use it as an io.ReadCloser.
type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() error {
	// we don't actually have to do anything here, since the buffer is just some data in memory
	// and the error is initialized to no-error.
	return nil
}

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
	// this is the multipart form buffer
	b := closingBuffer{new(bytes.Buffer)}

	appName := path.Base(appDir)
	log.Debugf("APP NAME: %s", appName)

	log.Debug("assembling build context archive")
	buildContext, err := tarBuildContext(appDir)
	if err != nil {
		return nil, err
	}

	log.Debug("assembling chart archive")
	chartTar, err := tarChart(path.Join(appDir, "chart"))
	if err != nil {
		return nil, err
	}

	// Prepare a form to upload the build context and chart archives.
	w := multipart.NewWriter(&b)
	buildContextFormFile, err := w.CreateFormFile("release-tar", "build.tar.gz")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(buildContextFormFile, buildContext); err != nil {
		return nil, err
	}

	// Add the other fields
	chartFormFile, err := w.CreateFormFile("chart-tar", "chart.tar.gz")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(chartFormFile, chartTar); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	c.Endpoint.Path = "/apps/" + appName
	req := &http.Request{
		Method:     "POST",
		URL:        c.Endpoint,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     c.Header,
		Body:       &b,
		Host:       c.Endpoint.Host,
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// NOTE(bacongobbler): after accepting the tarballs, the server upgrades the connection to
	// a websocket connection, so we need to switch protocols.
	if resp.StatusCode != http.StatusSwitchingProtocols {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("server responded with status code %d and failed to read the body: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("server responded with status code %d: %s", resp.StatusCode, body)
	}
	return nil, nil
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

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// tarBuildContext archives the given directory and returns the archive as an io.ReadCloser.
func tarBuildContext(path string) (io.ReadCloser, error) {
	pathExists, err := exists(path)
	if err != nil {
		return nil, err
	}

	if !pathExists {
		return nil, fmt.Errorf("directory '%s' does not exist", path)
	}

	options := archive.TarOptions{
		ExcludePatterns: []string{
			// do not include the chart directory. That will be packaged separately.
			"chart",
		},
		Compression: archive.Gzip,
	}
	return archive.TarWithOptions(path, &options)
}

// tarChart archives the directory and returns the archive as an io.ReadCloser.
func tarChart(path string) (io.ReadCloser, error) {
	pathExists, err := exists(path)
	if err != nil {
		return nil, err
	}

	if !pathExists {
		return nil, fmt.Errorf("chart directory '%s' does not exist", path)
	}

	return archive.Tar(path, archive.Gzip)
}
