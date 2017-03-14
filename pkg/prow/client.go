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
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
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

	return &Client{client, endpoint, false}
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
func (c Client) Up(appName, namespace string, out io.Writer, buildContext, chartReader io.ReadCloser, rawVals []byte) error {
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
	req.Body = &b
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Kubernetes-Namespace", namespace)
	req.Header.Set("Helm-Flag-Wait", strconv.FormatBool(c.OptionWait))
	req.Header.Set("Helm-Flag-Set", string(rawVals))

	log.Debugf("REQUEST: %s %s %s", req.Method, req.URL.String(), req.Header)

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
			}
			return err
		}
		io.Copy(out, p)
	}
}

// UpFromDir prepares the contents of appDir to create a build context and chart archive, then
// calls Up().
func (c Client) UpFromDir(appName, namespace string, out io.Writer, appDir string, rawVals []byte) error {

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

	return c.Up(appName, namespace, out, buildContext, chartTar, rawVals)
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
	contextDir, relDockerfile, err := builder.GetContextFromLocalDir(dir, "")

	if err != nil {
		return nil, fmt.Errorf("unable to prepare docker context: %s", err)
	}

	// canonicalize dockerfile name to a platform-independent one
	relDockerfile, err = archive.CanonicalTarNameForPath(relDockerfile)
	if err != nil {
		return nil, fmt.Errorf("cannot canonicalize dockerfile path %s: %v", relDockerfile, err)
	}

	f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer f.Close()

	var excludes []string
	if err == nil {
		excludes, err = dockerignore.ReadAll(f)
		if err != nil {
			return nil, err
		}
	}

	// do not include the chart directory. That will be packaged separately.
	excludes = append(excludes, filepath.Join(contextDir, "chart"))

	if err := builder.ValidateContextDirectory(contextDir, excludes); err != nil {
		return nil, fmt.Errorf("Error checking docker context: '%s'.", err)
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed. The daemon will remove them for us, if needed, after it
	// parses the Dockerfile. Ignore errors here, as they will have been
	// caught by validateContextDirectory above.
	var includes = []string{"."}
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", relDockerfile)
	}

	log.Debugf("INCLUDES: %v", includes)
	log.Debugf("EXCLUDES: %v", excludes)
	return archive.TarWithOptions(contextDir, &archive.TarOptions{
		Compression:     archive.Gzip,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	})
}

// tarChart archives the directory and returns the archive as an io.ReadCloser.
func tarChart(dir string) (io.ReadCloser, error) {
	dirExists, err := exists(filepath.Join(dir, "chart"))
	if err != nil {
		return nil, err
	}

	if !dirExists {
		return nil, ErrChartNotExist
	}

	options := archive.TarOptions{
		IncludeFiles: []string{
			"chart",
		},
		Compression: archive.Gzip,
	}

	return archive.TarWithOptions(dir, &options)
}
