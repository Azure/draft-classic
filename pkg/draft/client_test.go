package draft

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"

	"github.com/Azure/draft/api"
)

const expectedURLPath = "/apps/testdata"

type testWebsocketServerHandler struct{ *testing.T }
type testWebsocketServerHandlerCloseUnsupportedData struct{ *testing.T }

type testWebsocketServer struct {
	Server *httptest.Server
	URL    string
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func (t *testWebsocketServer) Close() {
	t.Server.Close()
}

func newTestWebsocketServer(t *testing.T) *testWebsocketServer {
	var s = new(testWebsocketServer)
	s.Server = httptest.NewServer(testWebsocketServerHandler{t})
	s.URL = makeWsProto(s.Server.URL)
	return s
}

func newTestWebsocketServerCloseUnsupportedData(t *testing.T) *testWebsocketServer {
	var s = new(testWebsocketServer)
	s.Server = httptest.NewServer(testWebsocketServerHandlerCloseUnsupportedData{t})
	s.URL = makeWsProto(s.Server.URL)
	return s
}

func (t testWebsocketServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != expectedURLPath {
		t.Logf("path=%v, want %v", r.URL.Path, expectedURLPath)
		http.Error(w, fmt.Sprintf("bad path %v, expected %v", r.URL.Path, expectedURLPath), http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(32 << 20)

	buildContext, _, err := r.FormFile("release-tar")
	if err != nil {
		t.Logf("no release-tar file found in multipart form")
		http.Error(w, fmt.Sprintf("error while reading release-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer buildContext.Close()

	chartFile, _, err := r.FormFile("chart-tar")
	if err != nil {
		t.Logf("no chart-tar file found in multipart form")
		http.Error(w, fmt.Sprintf("error while reading chart-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer chartFile.Close()

	ws, err := api.WebsocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		t.Logf("Upgrade: %v", err)
		return
	}
	defer ws.Close()

	ws.WriteMessage(websocket.TextMessage, []byte("hi there!"))

	ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
}

func (t testWebsocketServerHandlerCloseUnsupportedData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != expectedURLPath {
		t.Logf("path=%v, want %v", r.URL.Path, expectedURLPath)
		http.Error(w, fmt.Sprintf("bad path %v, expected %v", r.URL.Path, expectedURLPath), http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(32 << 20)

	buildContext, _, err := r.FormFile("release-tar")
	if err != nil {
		t.Logf("no release-tar file found in multipart form")
		http.Error(w, fmt.Sprintf("error while reading release-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer buildContext.Close()

	chartFile, _, err := r.FormFile("chart-tar")
	if err != nil {
		t.Logf("no chart-tar file found in multipart form")
		http.Error(w, fmt.Sprintf("error while reading chart-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer chartFile.Close()

	ws, err := api.WebsocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		t.Logf("Upgrade: %v", err)
		return
	}
	defer ws.Close()

	ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseUnsupportedData, ""),
		time.Now().Add(time.Second))
}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func TestNew(t *testing.T) {
	u, err := url.Parse("http://draft.rocks/foo?bar=car#star")
	if err != nil {
		t.Fatal(err)
	}
	client := New(u, nil)

	if client.HTTPClient == nil {
		t.Error("Excepted a default http client, got nil")
	}

	if client.Endpoint.Path != "" {
		t.Errorf("expected Path to be empty, got '%s'", client.Endpoint.Path)
	}

	if client.Endpoint.RawQuery != "" {
		t.Errorf("expected RawQuery to be empty, got '%s'", client.Endpoint.RawQuery)
	}

	if client.Endpoint.Fragment != "" {
		t.Errorf("expected Fragment to be empty, got '%s'", client.Endpoint.Fragment)
	}
}

func TestNewFromString(t *testing.T) {
	client, err := NewFromString("https://user:password@localhost/foo?bar=car#star", nil)
	if err != nil {
		t.Errorf("expected NewFromString to pass, got '%v'", err)
	}

	if client.HTTPClient == nil {
		t.Error("Excepted a default http client, got nil")
	}

	if client.Endpoint.Path != "" {
		t.Errorf("expected Path to be empty, got '%s'", client.Endpoint.Path)
	}

	if client.Endpoint.RawQuery != "" {
		t.Errorf("expected RawQuery to be empty, got '%s'", client.Endpoint.RawQuery)
	}

	if client.Endpoint.Fragment != "" {
		t.Errorf("expected Fragment to be empty, got '%s'", client.Endpoint.Fragment)
	}
}

func TestUpFromDir(t *testing.T) {
	ts := newTestWebsocketServer(t)
	defer ts.Close()

	client, err := NewFromString(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = client.UpFromDir("foo", "default", ioutil.Discard, "/ahsdkfjhaksdf", []byte{})
	if err == nil {
		t.Error("expected .UpFromDir() with invalid path to fail")
	}
	if err.Error() != "unable to prepare docker context: unable to evaluate symlinks in context path: lstat /ahsdkfjhaksdf: no such file or directory" {
		t.Errorf("expected .UpFromDir() with invalid path to fail as expected, got '%s'", err.Error())
	}

	if err := client.UpFromDir("testdata", "default", ioutil.Discard, "./testdata/good", []byte{}); err != nil {
		t.Errorf("expected .UpFromDir() with valid path to pass, got %v", err)
	}

	ts2 := newTestWebsocketServerCloseUnsupportedData(t)
	defer ts2.Close()

	client, err = NewFromString(ts2.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = client.UpFromDir("testdata", "default", ioutil.Discard, "./testdata/good", []byte{})
	if err == nil {
		t.Error("expected .UpFromDir() with bad server to fail")
	}
	if !websocket.IsCloseError(err, websocket.CloseUnsupportedData) {
		t.Errorf("expected err to be a CloseUnsupportedData error, got '%v'", err)
	}
}

func TestBadData(t *testing.T) {
	// don't care about setting up anything because we shouldn't hit the server.
	client := &Client{}

	if err := client.UpFromDir("testdata", "default", ioutil.Discard, "./testdata/no-dockerfile", []byte{}); err == nil {
		t.Error("expected .UpFromDir() with no Dockerfile to return an err")
	}

	if err := client.UpFromDir("testdata", "default", ioutil.Discard, "./testdata/no-chart", []byte{}); err != ErrChartNotExist {
		t.Errorf("expected .UpFromDir() with no Dockerfile to return ErrChartNotExist, got %v", err)
	}
}

func TestUpHeaders(t *testing.T) {
	var expectedNamespace = "testdata"
	var expectedValues = `aGVsbG86IHdvcmxkCmdvb2RieWU6IHdvcmxk`
	var expectedDecodedValues = "hello: world\ngoodbye: world"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Kubernetes-Namespace") != expectedNamespace {
			t.Errorf("expected Kubernetes-Namespace = %s, got %s", expectedNamespace, r.Header.Get("Kubernetes-Namespace"))
		}
		actualValues := r.Header.Get("Helm-Flag-Set")
		if actualValues != expectedValues {
			t.Errorf("expected Helm-Flag-Set = '%s', got '%s'", expectedValues, actualValues)
		}
		decodedVals, err := base64.StdEncoding.DecodeString(actualValues)
		if err != nil {
			t.Fatal(err)
		}
		if string(decodedVals) != expectedDecodedValues {
			t.Errorf("expected decoded Helm-Flag-Set = '%s', got '%s'", expectedDecodedValues, decodedVals)
		}
	}))
	defer ts.Close()

	client, err := NewFromString(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	client.UpFromDir("testdata", expectedNamespace, ioutil.Discard, "./testdata/good", []byte("hello: world\ngoodbye: world"))
}

func TestVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"semver": "v0.1.0", "git-commit": "abc123", "git-tree-state": "dirty"}`))
	}))
	defer ts.Close()
	client, err := NewFromString(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ver, err := client.Version()
	if err != nil {
		t.Errorf("expected no error to occur when retrieving server version, got '%v'", err)
	}
	if ver.SemVer != "v0.1.0" {
		t.Errorf("expected server semver to be v0.1.0, got '%s'", ver.SemVer)
	}
	if ver.GitCommit != "abc123" {
		t.Errorf("expected server semver to be abc123, got '%s'", ver.GitCommit)
	}
	if ver.GitTreeState != "dirty" {
		t.Errorf("expected server semver to be dirty, got '%s'", ver.GitTreeState)
	}
}

func TestTarBuildContext(t *testing.T) {
	diContext, err := tarBuildContext("./testdata/has-dockerignore")
	if err != nil {
		t.Fatal(err)
	}
	defer diContext.Close()
	diContents, err := ioutil.ReadAll(diContext)
	if err != nil {
		t.Fatal(err)
	}

	goodContext, err := tarBuildContext("./testdata/no-dockerignore")
	if err != nil {
		t.Fatal(err)
	}
	defer goodContext.Close()
	goodContents, err := ioutil.ReadAll(goodContext)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(diContents, goodContents) {
		t.Error("expected app with .dockerignore to differ from app without .dockerignore")
	}
}
