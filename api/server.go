package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	docker "github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/julienschmidt/httprouter"
)

// APIServer is an API Server which listens and responds to HTTP requests.
type APIServer struct {
	HTTPServer *http.Server
	Listener   net.Listener
	DockerClient *docker.Client
}

// Serve starts the HTTP server, accepting all new connections.
func (s *APIServer) Serve() error {
	return s.HTTPServer.Serve(s.Listener)
}

// Close shuts down the HTTP server, dropping all current connections.
func (s *APIServer) Close() error {
	return s.Listener.Close()
}

// ServeRequest processes a single HTTP request.
func (s *APIServer) ServeRequest(w http.ResponseWriter, req *http.Request) {
	s.HTTPServer.Handler.ServeHTTP(w, req)
}

func (s *APIServer) createRouter() {
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
			r.Handle(method, route, s.serverMiddleware(logRequestMiddleware(funct)))
		}
	}
	s.HTTPServer.Handler = r
}

func (s *APIServer) serverMiddleware(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// attach the API server to the request params so that it can retrieve info about itself
		ctx := context.WithValue(r.Context(), "server", s)
		// Delegate request to the given handle
		h(w, r.WithContext(ctx), p)
	}
}

// NewServer sets up the required Server and does protocol specific checking.
func NewServer(proto, addr string) (*APIServer, error) {
	var (
		a *APIServer
		err error
	)
	switch proto {
	case "tcp":
		a, err = setupTCPHTTP(addr)
	case "unix":
		a, err = setupUnixHTTP(addr)
	default:
		a, err = nil, fmt.Errorf("Invalid protocol format.")
	}
	a.createRouter()
	return a, err
}

func setupTCPHTTP(addr string) (*APIServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	a := &APIServer{
		HTTPServer: &http.Server{Addr: addr},
		Listener: l,
	}
	return a, nil
}

func setupUnixHTTP(addr string) (*APIServer, error) {
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

	a := &APIServer{
		HTTPServer: &http.Server{Addr: addr},
		Listener: l,
	}
	return a, nil
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
	imageName := fmt.Sprintf("127.0.0.1:5000/%s:latest", appName)
	server := r.Context().Value("server").(*APIServer)

	// TODO: check if app exists

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(32 << 20) // a maximum of 32MB is stored in memory before storing in temp files
	buildContext, _, err := r.FormFile("release-tar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	defer buildContext.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// send uploaded tar to docker as the build context
	buildResp, err := server.DockerClient.ImageBuild(
		context.Background(),
		buildContext,
		types.ImageBuildOptions{
			Tags: []string{imageName},
		})
	if err != nil {
		conn.Write([]byte(err.Error() + "\n"))
		return
	}
	defer buildResp.Body.Close()
	io.Copy(conn, buildResp.Body)
	pushResp, err := server.DockerClient.ImagePush(
		context.Background(),
		imageName,
		// assume no creds required for now
		// TODO(bacongobbler): implement custom auth handling for a registry
		types.ImagePushOptions{RegistryAuth: "foo"})
	if err != nil {
		conn.Write([]byte(err.Error() + "\n"))
		return
	}
	defer pushResp.Close()
	io.Copy(conn, pushResp)
	// TODO: install/upgrade via helm
}
