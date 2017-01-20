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
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/tiller/portforwarder"
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
	imageName := fmt.Sprintf("127.0.0.1:5000/%s:latest",
		appName,
	)
	server := r.Context().Value("server").(*APIServer)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// this is just a buffer of 32MB. Everything is piped over to docker's build context and to
	// Helm so this is just a sane default in case docker or helm's backed up somehow.
	r.ParseMultipartForm(32 << 20)
	buildContext, _, err := r.FormFile("release-tar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer buildContext.Close()

	chartFile, _, err := r.FormFile("chart-tar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer chartFile.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking.", http.StatusInternalServerError)
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
		fmt.Fprintf(conn, "!!! Could not build image from build context: %v\n", err)
		return
	}
	defer buildResp.Body.Close()
	io.Copy(conn, buildResp.Body)
	pushResp, err := server.DockerClient.ImagePush(
		context.Background(),
		imageName,
		// assume no creds required for now
		// TODO(bacongobbler): implement custom auth handling for a registry
		types.ImagePushOptions{RegistryAuth: "hi"})
	if err != nil {
		fmt.Fprintf(conn, "!!! Could not load push %s to registry: %v\n", imageName, err)
		return
	}
	defer pushResp.Close()
	io.Copy(conn, pushResp)

	tunnel, err := portforwarder.New("kube-system", "")
	if err != nil {
		fmt.Fprintf(conn, "!!! Could not get a connection to tiller: %v\n", err)
		return
	}

	client := helm.NewClient(helm.Host(fmt.Sprintf("localhost:%d", tunnel.Local)))
	chart, err := chartutil.LoadArchive(chartFile)
	if err != nil {
		fmt.Fprintf(conn, "!!! Could not load chart archive: %v\n", err)
		return
	}
	// inject certain values into the chart such as the registry location, the application name
	// and the version
	vals := fmt.Sprintf(`name="%s",version="%s",registry="%s:%s"`,
		appName,
		"latest",
		os.Getenv("PROWD_SERVICE_HOST"),
		os.Getenv("PROWD_SERVICE_PORT_REGISTRY"),
	)
	releaseResp, err := client.InstallReleaseFromChart(
		chart,
		appName,
		helm.ValueOverrides([]byte(vals)))
	if err != nil {
		fmt.Fprintf(conn, "!!! Could not install chart: %v\n", err)
		return
	}
	fmt.Fprintf(conn, "%s\n", releaseResp.Release.Info.Status.String())
}
