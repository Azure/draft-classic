package api

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller/portforwarder"

	"github.com/deis/prow/pkg/version"
)

const CloseBuildError = 3333

var WebsocketUpgrader = websocket.Upgrader{
	EnableCompression: true,
}

// APIServer is an API Server which listens and responds to HTTP requests.
type APIServer struct {
	HTTPServer   *http.Server
	Listener     net.Listener
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
			"/ping":    ping,
			"/version": getVersion,
		},
		"POST": {
			"/apps/:id": buildApp,
		},
	}

	for method, routes := range routerMap {
		for route, funct := range routes {
			// disable logging on /ping requests
			if route != "/ping" {
				funct = logRequestMiddleware(funct)
			}
			r.Handle(method, route, s.serverMiddleware(funct))
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
		a   *APIServer
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
		Listener:   l,
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
		Listener:   l,
	}
	return a, nil
}

func logRequestMiddleware(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Infof("%s %s", r.Method, r.RequestURI)
		// Delegate request to the given handle
		h(w, r, p)
	}
}

// writeJSON writes the value v to the http response stream as json with standard
// json encoding.
func writeJSON(w http.ResponseWriter, v interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func ping(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Write([]byte{'P', 'O', 'N', 'G'})
}

func getVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := writeJSON(w, version.New(), http.StatusOK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildApp(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	appName := p.ByName("id")
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
		http.Error(w, fmt.Sprintf("error while reading release-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer buildContext.Close()

	chartFile, _, err := r.FormFile("chart-tar")
	if err != nil {
		http.Error(w, fmt.Sprintf("error while reading chart-tar: %v", err), http.StatusBadRequest)
		return
	}
	defer chartFile.Close()

	conn, err := WebsocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error when upgrading connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// write build context to a buffer so we can also write to the sha1 hash
	buf := new(bytes.Buffer)
	buildContextChecksum := sha1.New()
	mw := io.MultiWriter(buf, buildContextChecksum)
	io.Copy(mw, buildContext)

	// truncate checksum to the first 40 characters (20 bytes)
	// this is the equivalent of `shasum build.tar.gz | awk '{print $1}'`
	tag := fmt.Sprintf("%.20x", buildContextChecksum.Sum(nil))
	imageName := fmt.Sprintf("127.0.0.1:5000/%s:%s",
		appName,
		tag,
	)

	// send uploaded tar to docker as the build context
	conn.WriteMessage(websocket.TextMessage, []byte("--> Building Dockerfile"))
	buildResp, err := server.DockerClient.ImageBuild(
		context.Background(),
		buf,
		types.ImageBuildOptions{
			Tags: []string{imageName},
		})
	if err != nil {
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				CloseBuildError,
				fmt.Sprintf("!!! Could not build image from build context: %v", err)))
		return
	}
	defer buildResp.Body.Close()
	if log.GetLevel() == log.DebugLevel {
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					CloseBuildError,
					fmt.Sprintf("There was an error fetching a text message writer: %v", err)))
		}
		io.Copy(w, buildResp.Body)
	}
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("--> Pushing %s", imageName)))
	pushResp, err := server.DockerClient.ImagePush(
		context.Background(),
		imageName,
		// assume no creds required for now
		// TODO(bacongobbler): implement custom auth handling for a registry
		types.ImagePushOptions{RegistryAuth: "hi"})
	if err != nil {
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				CloseBuildError,
				fmt.Sprintf("!!! Could not push %s to registry: %v", imageName, err)))
		return
	}
	defer pushResp.Close()
	if log.GetLevel() == log.DebugLevel {
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					CloseBuildError,
					fmt.Sprintf("There was an error fetching a text message writer: %v", err)))
		}
		io.Copy(w, pushResp)
	}

	conn.WriteMessage(websocket.TextMessage, []byte("--> Deploying to Kubernetes"))
	tunnel, err := portforwarder.New("kube-system", "")
	if err != nil {
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				CloseBuildError,
				fmt.Sprintf("!!! Could not get a connection to tiller: %v", err)))
		return
	}

	client := helm.NewClient(helm.Host(fmt.Sprintf("localhost:%d", tunnel.Local)))
	chart, err := chartutil.LoadArchive(chartFile)
	if err != nil {
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				CloseBuildError,
				fmt.Sprintf("!!! Could not load chart archive: %v", err)))
		return
	}
	// inject certain values into the chart such as the registry location, the application name
	// and the version
	vals := fmt.Sprintf("name: %s\nversion: %s\nregistry: \"%s:%s\"",
		appName,
		tag,
		os.Getenv("PROWD_SERVICE_HOST"),
		os.Getenv("PROWD_SERVICE_PORT_REGISTRY"),
	)
	// If a release does not exist, install it. If another error occurs during
	// the check, ignore the error and continue with the upgrade.
	//
	// The returned error is a grpc.rpcError that wraps the message from the original error.
	// So we're stuck doing string matching against the wrapped error, which is nested somewhere
	// inside of the grpc.rpcError message.
	_, err = client.ReleaseContent(appName, helm.ContentReleaseVersion(1))
	if err != nil && strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error()) {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte(fmt.Sprintf("    Release %q does not exist. Installing it now.", appName)))
		releaseResp, err := client.InstallReleaseFromChart(
			chart,
			"default",
			helm.ReleaseName(appName),
			helm.ValueOverrides([]byte(vals)))
		if err != nil {
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					CloseBuildError,
					fmt.Sprintf("!!! Could not install release: %v", err)))
			return
		}
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte(fmt.Sprintf("--> %s", releaseResp.Release.Info.Status.String())))
	} else {
		releaseResp, err := client.UpdateReleaseFromChart(
			appName,
			chart,
			helm.UpdateValueOverrides([]byte(vals)))
		if err != nil {
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					CloseBuildError,
					fmt.Sprintf("!!! Could not install release: %v", err)))
			return
		}
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte(fmt.Sprintf("--> %s", releaseResp.Release.Info.Status.String())))
	}

	// gently tell the client that we are closing the connection
	conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
}
