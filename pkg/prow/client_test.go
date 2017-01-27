package prow

import (
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
