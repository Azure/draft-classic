package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/kube"

	"github.com/deis/prow/pkg/prow"
	"github.com/deis/prow/pkg/prowd"
	"github.com/deis/prow/pkg/prowd/portforwarder"
)

const (
	hostEnvVar = "PROWD_HOST"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug   bool
	kubeContext string
	// prowdTunnel is a tunnelled connection used to send requests to prowd.
	// TODO refactor out this global var
	prowdTunnel *kube.Tunnel
	// prowdHost depicts where the prowd server is hosted.
	prowdHost string
)

var globalUsage = `The application deployment tool for Kubernetes.

Commands:

- prow create:    transform the local directory to be deployable via prow
- prow up:        packages and deploys your local directory to Kubernetes
- prow version:   display client version information
`

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "prow",
		Short:        "The application deployment tool for Kubernetes.",
		Long:         globalUsage,
		SilenceUsage: true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			teardown()
		},
	}
	p := cmd.PersistentFlags()
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	p.StringVar(&kubeContext, "kube-context", "", "name of the kubeconfig context to use")
	p.StringVar(&prowdHost, "host", defaultProwdHost(), "address of prowd. Overrides $PROWD_HOST")

	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	cmd.AddCommand(
		newCreateCmd(out),
		newUpCmd(out),
		newVersionCmd(out),
	)

	return cmd
}

func setupConnection(c *cobra.Command, args []string) error {
	if prowdHost == "" {
		tunnel, err := portforwarder.New(kubeContext)
		if err != nil {
			return err
		}

		prowdHost = fmt.Sprintf("http://localhost:%d", tunnel.Local)
		log.Debugf("Created tunnel using local port: '%d'\n", tunnel.Local)
	}

	log.Debugf("SERVER: %q\n", prowdHost)
	return nil
}

func teardown() {
	if prowdTunnel != nil {
		prowdTunnel.Close()
	}
}

func ensureProwClient(p prowd.Client) prowd.Client {
	if p != nil {
		return p
	}
	client, err := prow.NewFromString(prowdHost, nil)
	if err != nil {
		panic(err)
	}
	return client
}

func defaultProwdHost() string {
	return os.Getenv(hostEnvVar)
}

func main() {
	cmd := newRootCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
