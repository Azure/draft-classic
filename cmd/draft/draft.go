// Copyright (c) Microsoft Corporation. All rights reserved.
//
// Licensed under the MIT license.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/kube"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draftd/portforwarder"
)

const (
	hostEnvVar      = "DRAFT_HOST"
	homeEnvVar      = "DRAFT_HOME"
	namespaceEnvVar = "DRAFT_NAMESPACE"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug   bool
	kubeContext string
	// draftdTunnel is a tunnelled connection used to send requests to Draftd.
	// TODO refactor out this global var
	draftdTunnel *kube.Tunnel
	// draftHome depicts the home directory where all Draft config is stored.
	draftHome string
	// draftHost depicts where the Draftd server is hosted. This is used when the port forwarding logic by Kubernetes is unavailable.
	draftHost string
	// draftNamespace depicts which namespace the Draftd server is running in. This is used when Draftd was installed in a different namespace than kube-system.
	draftNamespace string
)

var globalUsage = `The application deployment tool for Kubernetes.
`

func newRootCmd(out io.Writer, in io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "draft",
		Short:        globalUsage,
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
	p.StringVar(&draftHome, "home", defaultDraftHome(), "location of your Draft config. Overrides $DRAFT_HOME")
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	p.StringVar(&kubeContext, "kube-context", "", "name of the kubeconfig context to use")
	p.StringVar(&draftHost, "host", defaultDraftHost(), "address of Draftd. This is used when the port forwarding feature by Kubernetes is unavailable. Overrides $DRAFT_HOST")
	p.StringVar(&draftNamespace, "draft-namespace", defaultDraftNamespace(), "namespace where Draftd is running. This is used when Draftd was installed in a different namespace than kube-system. Overrides $DRAFT_NAMESPACE")

	cmd.AddCommand(
		newCreateCmd(out),
		newHomeCmd(out),
		newInitCmd(out, in),
		newUpCmd(out),
		newVersionCmd(out),
		newPluginCmd(out),
		newConnectCmd(out),
		newDeleteCmd(out),
	)

	// Find and add plugins
	loadPlugins(cmd, draftpath.Home(homePath()), out, in)

	return cmd
}

func setupConnection(c *cobra.Command, args []string) error {
	if draftHost == "" {
		clientset, config, err := getKubeClient(kubeContext)
		if err != nil {
			return err
		}
		clientConfig, err := config.ClientConfig()
		if err != nil {
			return err
		}
		tunnel, err := portforwarder.New(clientset, clientConfig, draftNamespace)
		if err != nil {
			return err
		}

		draftHost = fmt.Sprintf("localhost:%d", tunnel.Local)
		log.Debugf("Created tunnel using local port: '%d'", tunnel.Local)
	}

	log.Debugf("SERVER: %q", draftHost)
	return nil
}

func teardown() {
	if draftdTunnel != nil {
		draftdTunnel.Close()
	}
}

func ensureDraftClient(client *draft.Client) *draft.Client {
	if client == nil {
		return draft.NewClient(&draft.ClientConfig{
			ServerAddr: draftHost,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
		})
	}
	return client
}

func defaultDraftHost() string {
	return os.Getenv(hostEnvVar)
}

func defaultDraftNamespace() string {
	if namespace := os.Getenv(namespaceEnvVar); namespace != "" {
		return namespace
	}
	return portforwarder.DefaultDraftNamespace
}

func defaultDraftHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".draft")
}

func homePath() string {
	return os.ExpandEnv(draftHome)
}

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*kubernetes.Clientset, clientcmd.ClientConfig, error) {
	config := kube.GetConfig(context)
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return client, config, nil
}

func main() {
	cmd := newRootCmd(os.Stdout, os.Stdin)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func debug(format string, args ...interface{}) {
	if flagDebug {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Printf(format, args...)
	}
}

func validateArgs(args, expectedArgs []string) error {
	if len(args) != len(expectedArgs) {
		return fmt.Errorf("This command needs %v argument(s): %v", len(expectedArgs), expectedArgs)
	}
	return nil
}
