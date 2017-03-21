package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"

	"github.com/deis/prow/pkg/prow"
	"github.com/deis/prow/pkg/prow/prowpath"
	"github.com/deis/prow/pkg/prowd/portforwarder"
)

const (
	hostEnvVar      = "PROW_HOST"
	homeEnvVar      = "PROW_HOME"
	namespaceEnvVar = "PROW_NAMESPACE"
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
	// prowNamespace depicts which kubernetes namespace the prowd server is hosted.
	prowNamespace string
)

var globalUsage = `The application deployment tool for Kubernetes.
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
	p.StringVar(&prowNamespace, "namespace", defaultProwNamespace(), "namespace of prowd. Overrides $PROW_NAMESPACE")

	cmd.AddCommand(
		newCreateCmd(out),
		newHomeCmd(out),
		newInitCmd(out),
		newUpCmd(out),
		newVersionCmd(out),
	)

	// Find and add plugins
	loadPlugins(cmd, prowpath.Home(homePath()), out)

	return cmd
}

func setupConnection(c *cobra.Command, args []string) error {
	if prowHost == "" {
		clientset, config, err := getKubeClient(kubeContext)
		if err != nil {
			return err
		}
		tunnel, err := portforwarder.New(prowNamespace, clientset, config)
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

func defaultProwNamespace() string {
	return os.Getenv(namespaceEnvVar)
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

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*internalclientset.Clientset, *restclient.Config, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return client, config, nil
}

func main() {
	cmd := newRootCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
