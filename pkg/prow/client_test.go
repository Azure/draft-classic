package prow

import (
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

	"github.com/deis/prow/api"
)

const expectedURLPath = "/apps/testdata"

type testWebsocketServerHandler struct{ *testing.T }

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
	var s *testWebsocketServer = new(testWebsocketServer)
	s.Server = httptest.NewServer(testWebsocketServerHandler{t})
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

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func TestNew(t *testing.T) {
	u, err := url.Parse("http://prow.rocks/foo?bar=car#star")
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

func TestUp(t *testing.T) {
	ts := newTestWebsocketServer(t)
	defer ts.Close()

	client, err := NewFromString(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Up("foo", "/ahsdkfjhaksdf", "default", ioutil.Discard)
	if err == nil {
		t.Error("expected .Up() with invalid path to fail")
	}
	if err.Error() != "directory '/ahsdkfjhaksdf' does not exist" {
		t.Errorf("expected .Up() with invalid path to fail as expected, got '%s'", err.Error())
	}

	if err := client.Up("testdata", "testdata", "default", ioutil.Discard); err != nil {
		t.Errorf("expected .Up() with valid path to pass, got %v", err)
	}

	// TODO(bacongobbler): write more extensive functional tests
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
