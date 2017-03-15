package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestBuildMiddleware(t *testing.T) {
	srv, err := NewServer("tcp", "0.0.0.0:4567")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	// create a custom router
	router := httprouter.New()
	router.Handle("GET", "/foo", srv.BuildMiddleware(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// do stuff
		time.Sleep(time.Millisecond)
	}))
	router.Handle("GET", "/bar", srv.BuildMiddleware(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// do stuff
		time.Sleep(time.Millisecond)
	}))
	srv.HTTPServer.Handler = router

	fooReq, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	barReq, err := http.NewRequest("GET", "/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	var numRequests = 20
	responseCodes := make(chan int, numRequests)
	for i := 0; i < numRequests/2; i++ {
		go getCode(srv, fooReq, responseCodes)
		go getCode(srv, barReq, responseCodes)
	}
	responseCodeCount := make(map[int]int)
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-responseCodes:
			responseCodeCount[code]++
		}
	}
	if responseCodeCount[http.StatusOK] != 2 {
		t.Errorf("expected there to be %dx '200 OK' responses, got %d", 2, responseCodeCount[http.StatusOK])
	}
	if responseCodeCount[http.StatusConflict] != numRequests-2 {
		t.Errorf("expected there to be %dx '409 CONFLICT' responses, got %d", numRequests-2, responseCodeCount[http.StatusConflict])
	}

	// sleep 1 millisecond so we guarantee that we pass the request sleep time, allowing us access to the resource
	time.Sleep(time.Millisecond)
	go getCode(srv, fooReq, responseCodes)
	go getCode(srv, barReq, responseCodes)

	for i := 0; i < 2; i++ {
		select {
		case code := <-responseCodes:
			responseCodeCount[code]++
		}
	}

	if responseCodeCount[http.StatusOK] != 4 {
		t.Errorf("expected there to be %dx '200 OK' responses, got %d", 4, responseCodeCount[http.StatusOK])
	}
}

func getCode(srv *APIServer, req *http.Request, responseCodeChan chan int) {
	r := httptest.NewRecorder()
	srv.ServeRequest(r, req)
	responseCodeChan <- r.Code
}
