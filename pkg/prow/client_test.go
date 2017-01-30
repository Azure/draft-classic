package prow

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name": "foo", "info": {"status": {"code": 1}}}`))
	}))
	defer ts.Close()

	client, err := NewFromString(ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Up("/foo", "default")
	if err == nil {
		t.Error("expected .Up() with invalid path to fail")
	}
	if err.Error() != "directory '/foo' does not exist" {
		t.Errorf("expected .Up() with invalid path to fail as expected, got '%s'", err.Error())
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
