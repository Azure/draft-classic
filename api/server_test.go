package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	srv, err := New("tcp", "0.0.0.0:4567")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	r := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.ServeRequest(r, req)
	if r.Code != http.StatusOK {
		t.Errorf("%d OK expected, received %d\n", http.StatusOK, r.Code)
	}
	message := r.Body.String()
	if message != "PONG" {
		t.Errorf("%s expected, received %s\n", "PONG", message)
	}
}
