package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	docker "github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/helm/prow/api"
)

const startDesc = `
Starts the prow server.
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
}

func newStartCmd(out io.Writer) *cobra.Command {
	sc := &startCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the prow server",
		Long:  startDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sc.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.listenAddr, "listen-addr", "l", "tcp://0.0.0.0:8080", "the address the server listens on")
	f.StringVarP(&sc.dockerAddr, "docker-addr", "", "unix:///var/run/docker.sock", "the address the docker engine listens on")
	f.StringVarP(&sc.dockerVersion, "docker-version", "", "", "the API version of the docker engine")
	f.BoolVarP(&sc.dockerFromEnv, "docker-from-env", "", false, "retrieve docker engine information from environment")

	return cmd
}

func (c *startCmd) run() error {
	var (
		dockerClient *docker.Client
		err error
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

	server, err := api.NewServer(protoAndAddr[0], protoAndAddr[1])
	if err != nil {
		return fmt.Errorf("failed to create server at %s: %v", c.listenAddr, err)
	}
	server.DockerClient = dockerClient
	log.Printf("server is now listening at %s", c.listenAddr)
	if err = server.Serve(); err != nil {
		return err
	}
	return nil
}
