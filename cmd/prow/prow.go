package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/kube"

	"github.com/deis/prow/pkg/prow"
	"github.com/deis/prow/pkg/prowd/portforwarder"
)

const (
	hostEnvVar = "PROW_HOST"
	homeEnvVar = "PROW_HOME"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug   bool
	kubeContext string
	// prowdTunnel is a tunnelled connection used to send requests to prowd.
	// TODO refactor out this global var
	prowdTunnel *kube.Tunnel
	// prowHome depicts the home directory where all prow config is stored.
	prowHome string
	// prowHost depicts where the prowd server is hosted.
	prowHost string
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flagDebug {
				log.SetLevel(log.DebugLevel)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			teardown()
		},
	}
	p := cmd.PersistentFlags()
	p.StringVar(&prowHome, "home", defaultProwHome(), "location of your Prow config. Overrides $PROW_HOME")
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	p.StringVar(&kubeContext, "kube-context", "", "name of the kubeconfig context to use")
	p.StringVar(&prowHost, "host", defaultProwHost(), "address of prowd. Overrides $PROW_HOST")

	cmd.AddCommand(
		newCreateCmd(out),
		newUpCmd(out),
		newVersionCmd(out),
	)

	return cmd
}

func setupConnection(c *cobra.Command, args []string) error {
	if prowHost == "" {
		tunnel, err := portforwarder.New(kubeContext)
		if err != nil {
			return err
		}

		prowHost = fmt.Sprintf("http://localhost:%d", tunnel.Local)
		log.Debugf("Created tunnel using local port: '%d'", tunnel.Local)
	}

	log.Debugf("SERVER: %q", prowHost)
	return nil
}

func teardown() {
	if prowdTunnel != nil {
		prowdTunnel.Close()
	}
}

func ensureProwClient(p *prow.Client) *prow.Client {
	if p != nil {
		return p
	}
	client, err := prow.NewFromString(prowHost, nil)
	if err != nil {
		panic(err)
	}
	return client
}

func defaultProwHost() string {
	return os.Getenv(hostEnvVar)
}

func defaultProwHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}
	return filepath.Join(os.Getenv("HOME"), ".prow")
}

func homePath() string {
	return os.ExpandEnv(prowHome)
}

func main() {
	cmd := newRootCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
