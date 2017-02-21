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
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/gorilla/websocket"

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
	// OptionWait specifies whether or not to wait for all resources to be ready on `prow up`
	OptionWait bool
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

	return &Client{client, endpoint, make(http.Header), false}
}

// NewFromString returns a new Client given a string URL and an optional client.
func NewFromString(endpoint string, client *http.Client) (*Client, error) {
	e, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return New(e, client), nil
}

// Up uploads the build context and chart to prowd, then writes messages from prowd to out.
// appName specifies the Helm release to create/update, and namespace specifies which namespace
// to deploy the application into.
func (c Client) Up(appName, namespace string, out io.Writer, buildContext, chartReader io.ReadCloser) error {
	// this is the multipart form buffer
	b := closingBuffer{new(bytes.Buffer)}

	defer buildContext.Close()
	defer chartReader.Close()

	log.Debugf("APP NAME: %s", appName)

	// Prepare a form to upload the build context and chart archives.
	w := multipart.NewWriter(&b)
	buildContextFormFile, err := w.CreateFormFile("release-tar", "build.tar.gz")
	if err != nil {
		return err
	}
	if _, err = io.Copy(buildContextFormFile, buildContext); err != nil {
		return err
	}

	// Add the other fields
	chartFormFile, err := w.CreateFormFile("chart-tar", "chart.tar.gz")
	if err != nil {
		return err
	}
	if _, err = io.Copy(chartFormFile, chartReader); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	c.Endpoint.Path = "/apps/" + appName
	// because this is a websocket connection, we must switch the protocol from http(s) to ws(s).
	if strings.Contains(c.Endpoint.Scheme, "http") {
		c.Endpoint.Scheme = "ws" + strings.TrimPrefix(c.Endpoint.Scheme, "http")
	}
	req := websocket.DefaultRequest(c.Endpoint)
	req.Method = "POST"
	req.Header = c.Header
	req.Body = &b
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Kubernetes-Namespace", namespace)
	req.Header.Set("Log-Level", log.GetLevel().String())
	req.Header.Set("Helm-Flag-Wait", strconv.FormatBool(c.OptionWait))

	log.Debugf("REQUEST: %s %s", req.Method, req.URL.String())

	conn, resp, err := websocket.DefaultDialer.Dial(req)
	if err == websocket.ErrBadHandshake {
		// let's do some digging to tell the user why the handshake failed
		p, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("there was an error while reading the response body from a bad websocket handshake: %v", err)
		}
		return fmt.Errorf("there was an error initiating a websocket handshake with the server: %d %s", resp.StatusCode, string(p))
	} else if err != nil {
		return fmt.Errorf("there was an error while dialing the server: %v", err)
	}
	defer conn.Close()

	for {
		_, p, err := conn.NextReader()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				// server closed the connection, so we're done!
				return nil
			} else {
				return err
			}
		} else {
			io.Copy(out, p)
		}
	}
}

// UpFromDir prepares the contents of appDir to create a build context and chart archive, then
// calls Up().
func (c Client) UpFromDir(appName, namespace string, out io.Writer, appDir string) error {

	log.Debug("assembling build context archive")
	buildContext, err := tarBuildContext(appDir)
	if err != nil {
		return err
	}

	log.Debug("assembling chart archive")
	chartTar, err := tarChart(appDir)
	if err != nil {
		return err
	}

	return c.Up(appName, namespace, out, buildContext, chartTar)
}

// Version returns the server version.
func (c *Client) Version() (*version.Version, error) {
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
func exists(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// tarBuildContext archives the given directory and returns the archive as an io.ReadCloser.
func tarBuildContext(dir string) (io.ReadCloser, error) {
	dirExists, err := exists(dir)
	if err != nil {
		return nil, err
	}

	if !dirExists {
		return nil, fmt.Errorf("directory '%s' does not exist", dir)
	}

	dockerfileExists, err := exists(path.Join(dir, "Dockerfile"))
	if err != nil {
		return nil, err
	}

	if !dockerfileExists {
		return nil, DockerfileNotExistError
	}

	options := archive.TarOptions{
		ExcludePatterns: []string{
			// do not include the chart directory. That will be packaged separately.
			"chart",
		},
		Compression: archive.Gzip,
	}
	return archive.TarWithOptions(dir, &options)
}

// tarChart archives the directory and returns the archive as an io.ReadCloser.
func tarChart(dir string) (io.ReadCloser, error) {
	dirExists, err := exists(path.Join(dir, "chart"))
	if err != nil {
		return nil, err
	}

	if !dirExists {
		return nil, ChartNotExistError
	}

	options := archive.TarOptions{
		IncludeFiles: []string{
			"chart",
		},
		Compression: archive.Gzip,
	}

	return archive.TarWithOptions(dir, &options)
}
