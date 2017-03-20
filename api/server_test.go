package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestPing(t *testing.T) {
	srv, err := NewServer("tcp", "0.0.0.0:4567")
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

func TestServerMiddleware(t *testing.T) {
	srv, err := NewServer("tcp", "0.0.0.0:4567")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	router := httprouter.New()
	router.Handle("GET", "/", srv.Middleware(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		server := r.Context().Value(contextKey("server")).(*Server)
		if server != srv {
			t.Errorf("expected %v, got %v", srv, server)
		}
	}))
	srv.HTTPServer.Handler = router

	if _, err := http.NewRequest("GET", "/", nil); err != nil {
		t.Fatal(err)
	}
}
