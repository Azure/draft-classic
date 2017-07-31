package draft

import (
	"fmt"
	"github.com/Azure/draft/pkg/rpc"
	"golang.org/x/net/context"
	"net"
	"sync"
)

// kubernetes imports
import (
	// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	// 	"k8s.io/client-go/pkg/api/v1"
)

// docker imports
import (
	docker "github.com/docker/docker/client"
	//"github.com/docker/docker/pkg/jsonmessage"
	//"github.com/docker/docker/api/types"
	//"github.com/docker/docker/pkg/term"
)

// helm imports
import (
	//"k8s.io/helm/pkg/proto/hapi/release"
	//"k8s.io/helm/pkg/chartutil"
	//"k8s.io/helm/pkg/strvals"
	"k8s.io/helm/pkg/helm"
)

type (
	// RegistryConfig specifies configuration for the image repository.
	RegistryConfig struct {
		// Auth is the authorization token used to push images up to the registry.
		Auth string
		// Org is the organization (e.g. your DockerHub account) used to push images
		// up to the registry.
		Org string
		// URL is the URL of the registry (e.g. quay.io, docker.io, gcr.io)
		URL string
	}

	// RegistryAuth is the registry authentication credentials
	RegistryAuth struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		Email         string `json:"email"`
		RegistryToken string `json:"registrytoken"`
	}

	// DockerAuth is a container for the registry authentication credentials wrapped
	// by the registry server name.
	DockerAuth map[string]RegistryAuth
)

// ServerConfig specifies draft.Server configuration.
type ServerConfig struct {
	ListenAddr string
	Basedomain string // Basedomain is the basedomain used to construct the ingress rules
	Registry   *RegistryConfig
	Docker     *docker.Client
	Helm       helm.Interface
	Kube       *k8s.Clientset
}

// Server is a draft Server.
type Server struct {
	cfg *ServerConfig
	srv rpc.Server
}

// NewServer returns a draft.Server initialized with the
// provided configuration.
func NewServer(cfg *ServerConfig) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	errc := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.srv = rpc.NewServer()
		errc <- s.srv.Serve(lis, s)
		close(errc)
		close(done)
	}()
	select {
	case <-ctx.Done():
		s.srv.Stop()
		close(done)
		return ctx.Err()
	case <-done:
		return <-errc
	}
}

// Up handles incoming draft up requests and returns a stream of summaries or error.
//
// Up implements rpc.UpHandler
func (s *Server) Up(ctx context.Context, req *rpc.UpRequest) <-chan *rpc.UpSummary {
	ch := make(chan *rpc.UpSummary)
	go func() {
		fmt.Printf("NAMESPACE: %v\n", req.Namespace)
		fmt.Printf("CHART:     %v\n", req.Chart)
		fmt.Printf("VALUES:    %v\n", req.Values)
		fmt.Printf("FILES:     %v\n", req.Files)
		ch <- &rpc.UpSummary{StageName: "test", StatusText: "OK", StatusCode: rpc.UpSummary_SUCCESS}
		close(ch)
	}()
	return ch
}
