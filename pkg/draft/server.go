package draft

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/Azure/draft/pkg/rpc"
	"github.com/Azure/draft/pkg/storage"
)

const draftLogsDirPrefix = "draft-logs"

// ServerConfig specifies draft.Server configuration.
type ServerConfig struct {
	ListenAddr     string
	IngressEnabled bool
	Basedomain     string // Basedomain is the basedomain used to construct the ingress rules
	Registry       *RegistryConfig
	Docker         *docker.Client
	Helm           helm.Interface
	Kube           k8s.Interface
	UseTLS         bool
	TLSConfig      *tls.Config
	Storage        storage.Store
}

// Server is a draft Server.
type Server struct {
	cfg     *ServerConfig
	srv     rpc.Server
	logsDir string
}

// NewServer returns a draft.Server initialized with the
// provided configuration.
func NewServer(cfg *ServerConfig) *Server {
	return &Server{cfg: cfg}
}

// Serve starts draftd
func (s *Server) Serve(ctx context.Context) error {
	// create temporary logs directory
	var err error
	if s.logsDir, err = ioutil.TempDir("", draftLogsDirPrefix); err != nil {
		return fmt.Errorf("could not create logs directory: %v", err)
	}
	defer os.RemoveAll(s.logsDir)

	// start probes server
	cancelctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := s.probes(cancelctx, &wg); err != nil {
			fmt.Printf("error: probes: %v\n", err)
		}
		wg.Done()
	}()

	lis, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}

	var opts []rpc.ServerOpt
	if s.cfg.UseTLS {
		opts = append(opts, rpc.WithGrpcServerOpt(
			grpc.Creds(credentials.NewTLS(s.cfg.TLSConfig)),
		))
	}

	errc := make(chan error, 1)
	s.srv = rpc.NewServer(opts...)
	wg.Add(1)
	go func() {
		errc <- s.srv.Serve(lis, s)
		close(errc)
		wg.Done()
	}()

	defer func() {
		s.srv.Stop()
		cancel()
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errc:
		return err
	}
}

// regular expression to sanitize build ids.
var reBuildID = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// Logs handles incoming requests to retrieve logs for a draft build.
//
// Logs implements rpc.LogsHandler
func (s *Server) Logs(ctx context.Context, req *rpc.GetLogsRequest) (*rpc.GetLogsResponse, error) {
	if !reBuildID.MatchString(req.BuildID) {
		return nil, fmt.Errorf("invalid build id %q", req.BuildID)
	}
	obj, err := s.cfg.Storage.GetBuild(context.Background(), req.AppName, req.BuildID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve application (%q) storage object for build %q: %v",
			req.AppName,
			req.BuildID,
			err,
		)
	}
	var buf bytes.Buffer
	if err := tail(&buf, req.Limit, obj.LogsFileRef); err != nil {
		return nil, err
	}
	return &rpc.GetLogsResponse{Content: buf.Bytes()}, nil
}

// Up handles incoming draft up requests and returns a stream of summaries or error.
//
// Up implements rpc.UpHandler
func (s *Server) Up(ctx context.Context, req *rpc.UpRequest) <-chan *rpc.UpSummary {
	return s.buildApp(ctx, req)
}

func (s *Server) buildApp(ctx context.Context, req *rpc.UpRequest) <-chan *rpc.UpSummary {
	ch := make(chan *rpc.UpSummary, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var (
			buf = new(bytes.Buffer)
			app *AppContext
			err error
		)
		defer func() {
			s.finish(app, buf)
			wg.Done()
		}()
		w := io.MultiWriter(buf, os.Stdout, os.Stderr)
		if app, err = newAppContext(s, req, w); err != nil {
			fmt.Printf("buildApp: error creating app context: %v\n", err)
			return
		}
		if err := s.buildImg(ctx, app, ch); err != nil {
			fmt.Printf("buildApp: buildImg error: %v\n", err)
			return
		}
		if err := s.pushImg(ctx, app, ch); err != nil {
			fmt.Printf("buildApp: pushImg error: %v\n", err)
			return
		}
		if err := s.release(ctx, app, ch); err != nil {
			fmt.Printf("buildApp: release error: %v\n", err)
			return
		}
	}()
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// finish updates storage with the information collected during the stages of a draft build and
// writes the aggregated logs to a tempoarary file.
func (s *Server) finish(app *AppContext, buf *bytes.Buffer) {
	logsFile := filepath.Join(s.logsDir, app.id)
	app.obj.LogsFileRef = logsFile
	if err := s.cfg.Storage.UpdateBuild(context.Background(), app.req.AppName, app.obj); err != nil {
		fmt.Printf("complete: failed to store build object for app %q: %v\n", app.req.AppName, err)
		return
	}

	if err := ioutil.WriteFile(logsFile, buf.Bytes(), 0666); err != nil {
		fmt.Printf("complete: failed to write logs to file for build %q: %v\n", app.id, err)
		return
	}
	fmt.Printf("complete: wrote logs to %s\n", logsFile)
}

// buildImg builds the docker image.
func (s *Server) buildImg(ctx context.Context, app *AppContext, out chan<- *rpc.UpSummary) (err error) {
	const stageDesc = "Building Docker Image"

	defer complete(app.id, stageDesc, out, &err)
	summary := summarize(app.id, stageDesc, out)

	// notify that particular stage has started.
	summary("started", rpc.UpSummary_STARTED)

	msgc := make(chan string)
	errc := make(chan error)
	go func() {
		buildopts := types.ImageBuildOptions{Tags: []string{app.img}}
		resp, err := s.cfg.Docker.ImageBuild(ctx, app.buf, buildopts)
		if err != nil {
			errc <- err
			return
		}
		defer func() {
			resp.Body.Close()
			close(msgc)
			close(errc)
		}()
		outFd, isTerm := term.GetFdInfo(app.out)
		if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, app.out, outFd, isTerm, nil); err != nil {
			errc <- err
			return
		}
		if _, _, err = s.cfg.Docker.ImageInspectWithRaw(ctx, app.img); err != nil {
			if docker.IsErrImageNotFound(err) {
				errc <- fmt.Errorf("Could not locate image for %s: %v", app.req.AppName, err)
				return
			}
			errc <- fmt.Errorf("ImageInspectWithRaw error: %v", err)
			return
		}
	}()
	for msgc != nil || errc != nil {
		select {
		case msg, ok := <-msgc:
			if !ok {
				msgc = nil
				continue
			}
			summary(msg, rpc.UpSummary_LOGGING)
		case err, ok := <-errc:
			if !ok {
				errc = nil
				continue
			}
			return err
		default:
			summary("ongoing", rpc.UpSummary_ONGOING)
			time.Sleep(time.Second)
		}
	}
	return nil
}

// pushImg pushes the results of buildImg to the image repository.
func (s *Server) pushImg(ctx context.Context, app *AppContext, out chan<- *rpc.UpSummary) (err error) {
	const stageDesc = "Pushing Docker Image"

	defer complete(app.id, stageDesc, out, &err)
	summary := summarize(app.id, stageDesc, out)

	// notify that particular stage has started.
	summary("started", rpc.UpSummary_STARTED)

	msgc := make(chan string, 1)
	errc := make(chan error, 1)
	go func() {
		pushopts := types.ImagePushOptions{RegistryAuth: s.cfg.Registry.Auth}
		resp, err := s.cfg.Docker.ImagePush(ctx, app.img, pushopts)
		if err != nil {
			errc <- err
			return
		}
		defer func() {
			resp.Close()
			close(errc)
			close(msgc)
		}()
		outFd, isTerm := term.GetFdInfo(app.out)
		if err := jsonmessage.DisplayJSONMessagesStream(resp, app.out, outFd, isTerm, nil); err != nil {
			errc <- err
			return
		}
	}()
	for msgc != nil || errc != nil {
		select {
		case msg, ok := <-msgc:
			if !ok {
				msgc = nil
				continue
			}
			summary(msg, rpc.UpSummary_LOGGING)
		case err, ok := <-errc:
			if !ok {
				errc = nil
				continue
			}
			return err
		default:
			summary("ongoing", rpc.UpSummary_ONGOING)
			time.Sleep(time.Second)
		}
	}
	return nil
}

// release installs or updates the application deployment.
func (s *Server) release(ctx context.Context, app *AppContext, out chan<- *rpc.UpSummary) (err error) {
	const stageDesc = "Releasing Application"

	defer complete(app.id, stageDesc, out, &err)
	summary := summarize(app.id, stageDesc, out)

	// notify that particular stage has started.
	summary("started", rpc.UpSummary_STARTED)

	if err := s.prepareReleaseEnvironment(app); err != nil {
		return err
	}

	// If a release does not exist, install it. If another error occurs during the check,
	// ignore the error and continue with the upgrade.
	//
	// The returned error is a grpc.rpcError that wraps the message from the original error.
	// So we're stuck doing string matching against the wrapped error, which is nested inside
	// of the grpc.rpcError message.
	_, err = s.cfg.Helm.ReleaseContent(app.req.AppName, helm.ContentReleaseVersion(1))
	if err != nil && strings.Contains(err.Error(), "not found") {
		msg := fmt.Sprintf("Release %q does not exist. Installing it now.", app.req.AppName)
		summary(msg, rpc.UpSummary_LOGGING)

		vals, err := app.vals.YAML()
		if err != nil {
			return err
		}

		opts := []helm.InstallOption{
			helm.ReleaseName(app.req.AppName),
			helm.ValueOverrides([]byte(vals)),
			helm.InstallWait(app.req.GetOptions().GetReleaseWait()),
		}
		rls, err := s.cfg.Helm.InstallReleaseFromChart(app.req.Chart, app.req.Namespace, opts...)
		if err != nil {
			return fmt.Errorf("could not install release: %v", grpcError(err))
		}
		formatReleaseStatus(app, rls.Release, summary)

	} else {
		msg := fmt.Sprintf("Upgrading %s.", app.req.AppName)
		summary(msg, rpc.UpSummary_LOGGING)

		vals, err := app.vals.YAML()
		if err != nil {
			return err
		}

		opts := []helm.UpdateOption{
			helm.UpdateValueOverrides([]byte(vals)),
			helm.UpgradeWait(app.req.GetOptions().GetReleaseWait()),
		}
		rls, err := s.cfg.Helm.UpdateReleaseFromChart(app.req.AppName, app.req.Chart, opts...)
		if err != nil {
			return fmt.Errorf("could not upgrade release: %v", grpcError(err))
		}
		app.obj.Release = rls.Release.Name
		formatReleaseStatus(app, rls.Release, summary)
	}
	return nil
}

// probes starts a http server to handle livenes and readiness probes.
func (s *Server) probes(ctx context.Context, wg *sync.WaitGroup) error {
	const addr = ":8080"

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &http.Server{Addr: addr, Handler: http.HandlerFunc(s.health)}
	errc := make(chan error, 1)

	wg.Add(1)
	go func() {
		errc <- srv.Serve(lis)
		close(errc)
		wg.Done()
	}()
	defer func() {
		srv.Shutdown(ctx)
		lis.Close()
	}()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) prepareReleaseEnvironment(app *AppContext) error {
	// determine if the destination namespace exists, create it if not.
	if _, err := s.cfg.Kube.CoreV1().Namespaces().Get(app.req.Namespace, metav1.GetOptions{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
		_, err = s.cfg.Kube.CoreV1().Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: app.req.Namespace},
		})
		if err != nil {
			return fmt.Errorf("could not create namespace %q: %v", app.req.Namespace, err)
		}
	}

	regauth, err := configureRegistryAuth(s.cfg.Registry.Auth)
	if err != nil {
		return err
	}
	// create a new json string with the full dockerauth, including the registry URL.
	js, err := json.Marshal(DockerAuth{s.cfg.Registry.URL: regauth})
	if err != nil {
		return fmt.Errorf("could not json encode docker authentication string: %v", err)
	}

	// determine if the registry pull secret exists in the desired namespace, create it if not.
	var secret *v1.Secret
	if secret, err = s.cfg.Kube.CoreV1().Secrets(app.req.Namespace).Get(pullSecretName, metav1.GetOptions{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
		_, err = s.cfg.Kube.CoreV1().Secrets(app.req.Namespace).Create(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pullSecretName,
					Namespace: app.req.Namespace,
				},
				Type: v1.SecretTypeDockercfg,
				StringData: map[string]string{
					".dockercfg": string(js),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("could not create registry pull secret: %v", err)
		}
	} else {
		// the registry pull secret exists, check if it needs to be updated.
		if data, ok := secret.StringData[".dockercfg"]; ok && data != string(js) {
			secret.StringData[".dockercfg"] = string(js)
			_, err = s.cfg.Kube.CoreV1().Secrets(app.req.Namespace).Update(secret)
			if err != nil {
				return fmt.Errorf("could not update registry pull secret: %v", err)
			}
		}
	}

	// determine if the default service account in the desired namespace has the correct
	// imagePullSecret. If not, add it.
	svcAcct, err := s.cfg.Kube.CoreV1().ServiceAccounts(app.req.Namespace).Get(svcAcctNameDefault, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not load default service account: %v", err)
	}
	found := false
	for _, ps := range svcAcct.ImagePullSecrets {
		if ps.Name == pullSecretName {
			found = true
			break
		}
	}
	if !found {
		svcAcct.ImagePullSecrets = append(svcAcct.ImagePullSecrets, v1.LocalObjectReference{
			Name: pullSecretName,
		})
		_, err := s.cfg.Kube.CoreV1().ServiceAccounts(app.req.Namespace).Update(svcAcct)
		if err != nil {
			return fmt.Errorf("could not modify default service account with registry pull secret: %v", err)
		}
	}

	return nil
}

// health serves and responds to liveness and readiness probes.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func formatReleaseStatus(app *AppContext, rls *release.Release, summary func(string, rpc.UpSummary_StatusCode)) {
	status := fmt.Sprintf("%s %v", app.req.AppName, rls.Info.Status.Code)
	summary(status, rpc.UpSummary_LOGGING)
	if rls.Info.Status.Notes != "" {
		notes := fmt.Sprintf("notes: %v", rls.Info.Status.Notes)
		summary(notes, rpc.UpSummary_LOGGING)
	}
}

// summarize returns a function closure that wraps writing rpc.UpSummary_StatusCode.
func summarize(id, desc string, out chan<- *rpc.UpSummary) func(string, rpc.UpSummary_StatusCode) {
	return func(info string, code rpc.UpSummary_StatusCode) {
		out <- &rpc.UpSummary{StageDesc: desc, StatusText: info, StatusCode: code, BuildId: id}
	}
}

// complete marks the end of a draft build stage.
func complete(id, desc string, out chan<- *rpc.UpSummary, err *error) {
	switch fn := summarize(id, desc, out); {
	case *err != nil:
		fn(fmt.Sprintf("failure: %v", *err), rpc.UpSummary_FAILURE)
	default:
		fn("success", rpc.UpSummary_SUCCESS)
	}
}

func grpcError(err error) error {
	return errors.New(grpc.ErrorDesc(err))
}

// tail writes up to limit number of lines of the file specified by path relative to the file end.
func tail(buf *bytes.Buffer, limit int64, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		offset = int64(-1)
		nlines = int64(0)
		endPos int64
	)
	for nlines <= limit-1 {
		begPos, err := f.Seek(offset, 2)
		if err != nil {
			return err
		}
		if begPos == 0 {
			endPos = -1
			break
		}
		b := make([]byte, 1)
		if _, err = f.ReadAt(b, begPos); err != nil {
			return err
		}
		if offset == int64(-1) && string(b) == "\n" {
			offset--
			continue
		}
		if string(b) == "\n" {
			nlines++
			endPos = begPos
		}
		offset--
	}

	if _, err = f.Seek(endPos+1, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		buf.Write(scanner.Bytes())
		buf.WriteString("\n")
	}
	return nil
}
