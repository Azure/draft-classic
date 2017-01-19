package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/julienschmidt/httprouter"
)

// HTTPServer is an API Server which listens and responds to HTTP requests.
type HTTPServer struct {
	srv *http.Server
	l   net.Listener
}

// Serve starts the HTTP server, accepting all new connections.
func (s *HTTPServer) Serve() error {
	return s.srv.Serve(s.l)
}

// Close shuts down the HTTP server, dropping all current connections.
func (s *HTTPServer) Close() error {
	return s.l.Close()
}

// ServeRequest processes a single HTTP request.
func (s *HTTPServer) ServeRequest(w http.ResponseWriter, req *http.Request) {
	s.srv.Handler.ServeHTTP(w, req)
}

// NewServer sets up the required Server and does protocol specific checking.
func NewServer(proto, addr string) (*HTTPServer, error) {
	switch proto {
	case "tcp":
		return setupTCPHTTP(addr)
	case "unix":
		return setupUnixHTTP(addr)
	default:
		return nil, fmt.Errorf("Invalid protocol format.")
	}
}

func setupTCPHTTP(addr string) (*HTTPServer, error) {
	r := createRouter()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &HTTPServer{&http.Server{Addr: addr, Handler: r}, l}, nil
}

func setupUnixHTTP(addr string) (*HTTPServer, error) {
	r := createRouter()

	if err := syscall.Unlink(addr); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mask := syscall.Umask(0777)
	defer syscall.Umask(mask)

	l, err := net.Listen("unix", addr)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(addr, 0660); err != nil {
		return nil, err
	}

	return &HTTPServer{&http.Server{Addr: addr, Handler: r}, l}, nil
}

func createRouter() *httprouter.Router {
	r := httprouter.New()

	routerMap := map[string]map[string]httprouter.Handle{
		"GET": {
			"/ping": ping,
		},
		"POST": {
			"/apps/:id": buildApp,
		},
	}

	for method, routes := range routerMap {
		for route, funct := range routes {
			r.Handle(method, route, logRequestMiddleware(funct))
		}
	}
	return r
}

func logRequestMiddleware(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		// Delegate request to the given handle
		h(w, r, p)
	}
}

// WriteJSON writes the value v to the http response stream as json with standard
// json encoding.
func WriteJSON(w http.ResponseWriter, v interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func ping(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Write([]byte{'P', 'O', 'N', 'G'})
}

func buildApp(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	appName := p.ByName("id")

	// TODO: check if app exists

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(32 << 20) // a maximum of 32MB is stored in memory before storing in temp files
	file, _, err := r.FormFile("release-tar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()
	f, err := ioutil.TempFile("", appName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not create tmpdir to store tarball: " + err.Error()))
		return
	}
	defer f.Close()
	io.Copy(f, file)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// TODO: untar archive and run docker build on it
	// TODO: push to local registry
	// TODO: install/upgrade via helm
	bufrw.ReadString('\n')
}
