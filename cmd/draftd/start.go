package main

import (
	"fmt"
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
	docker "github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"

	"github.com/Azure/draft/api"
)

const startDesc = `
Starts the draft server.
`

type startCmd struct {
	out io.Writer
	// listenAddr is the address which the server will be listening on.
	listenAddr string
	// dockerAddr is the address which the docker engine listens on.
	dockerAddr string
	// dockerVersion is the API version of the docker engine. If unset, no version information is
	// sent to the engine, however it is strongly recommended by Docker to set this or the client
	// may break if the server is upgraded.
	dockerVersion string
	// retrieve docker engine information from environment
	dockerFromEnv bool
	// registryAuth is the authorization token used to push images up to the registry.
	registryAuth string
	// registryOrg is the organization (e.g. your DockerHub account) used to push images up to the registry.
	registryOrg string
	// registryURL is the URL of the registry (e.g. quay.io, docker.io, gcr.io)
	registryURL string
	// basedomain is the base domain used to construct the ingress host name to applications.
	basedomain string
	// tillerURI is the URI used to connect to tiller.
	tillerURI string
}

func newStartCmd(out io.Writer) *cobra.Command {
	sc := &startCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the draft server",
		Long:  startDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sc.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.listenAddr, "listen-addr", "l", "tcp://0.0.0.0:44135", "the address the server listens on")
	f.StringVarP(&sc.dockerAddr, "docker-addr", "", "unix:///var/run/docker.sock", "the address the docker engine listens on")
	f.StringVarP(&sc.dockerVersion, "docker-version", "", "", "the API version of the docker engine")
	f.BoolVarP(&sc.dockerFromEnv, "docker-from-env", "", true, "retrieve docker engine information from environment")
	f.StringVar(&sc.registryAuth, "registry-auth", "", "the authorization token used to push images up to the registry")
	f.StringVar(&sc.registryOrg, "registry-org", "", "the organization (e.g. your DockerHub account) used to push images up to the registry")
	f.StringVar(&sc.registryURL, "registry-url", "127.0.0.1:5000", "the URL of the registry (e.g. quay.io, docker.io, gcr.io)")
	f.StringVar(&sc.basedomain, "basedomain", "", "the base domain in which a wildcard DNS entry points to an ingress controller")
	f.StringVar(&sc.tillerURI, "tiller-uri", "tiller-deploy:44134", "the URI used to connect to tiller")

	return cmd
}

func (c *startCmd) run() error {
	var (
		dockerClient *docker.Client
		err          error
	)

	protoAndAddr := strings.SplitN(c.listenAddr, "://", 2)

	if c.dockerFromEnv {
		dockerClient, err = docker.NewEnvClient()
	} else {
		dockerClient, err = docker.NewClient(c.dockerAddr, c.dockerVersion, nil, nil)
	}
	if err != nil {
		return err
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	server, err := api.NewServer(protoAndAddr[0], protoAndAddr[1])
	if err != nil {
		return fmt.Errorf("failed to create server at %s: %v", c.listenAddr, err)
	}
	server.DockerClient = dockerClient
	server.RegistryAuth = c.registryAuth
	server.RegistryOrg = c.registryOrg
	server.RegistryURL = c.registryURL
	server.Basedomain = c.basedomain
	server.HelmClient = helm.NewClient(helm.Host(c.tillerURI))
	server.KubeClient = kubeClientset
	log.Printf("server is now listening at %s", c.listenAddr)
	return server.Serve()
}
