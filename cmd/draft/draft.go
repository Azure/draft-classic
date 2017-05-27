// Copyright (c) Microsoft Corporation. All rights reserved.
//
// Licensed under the MIT license.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draftd/portforwarder"
)

const (
	hostEnvVar = "DRAFT_HOST"
	homeEnvVar = "DRAFT_HOME"
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
)

var globalUsage = `The application deployment tool for Kubernetes.
`

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "draft",
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
	p.StringVar(&draftHome, "home", defaultDraftHome(), "location of your Draft config. Overrides $DRAFT_HOME")
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	p.StringVar(&kubeContext, "kube-context", "", "name of the kubeconfig context to use")
	p.StringVar(&draftHost, "host", defaultDraftHost(), "address of Draftd. Overrides $DRAFT_HOST")

	cmd.AddCommand(
		newCreateCmd(out),
		newHomeCmd(out),
		newInitCmd(out),
		newUpCmd(out),
		newVersionCmd(out),
	)

	// Find and add plugins
	loadPlugins(cmd, draftpath.Home(homePath()), out)

	return cmd
}

func setupConnection(c *cobra.Command, args []string) error {
	if draftHost == "" {
		clientset, config, err := getKubeClient(kubeContext)
		if err != nil {
			return err
		}
		tunnel, err := portforwarder.New(clientset, config)
		if err != nil {
			return err
		}

		draftHost = fmt.Sprintf("http://localhost:%d", tunnel.Local)
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

func ensureDraftClient(p *draft.Client) *draft.Client {
	if p != nil {
		return p
	}
	client, err := draft.NewFromString(draftHost, nil)
	if err != nil {
		panic(err)
	}
	return client
}

func defaultDraftHost() string {
	return os.Getenv(hostEnvVar)
}

func defaultDraftHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}
	return filepath.Join(os.Getenv("HOME"), ".draft")
}

func homePath() string {
	return os.ExpandEnv(draftHome)
}

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*kubernetes.Clientset, *restclient.Config, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := kubernetes.NewForConfig(config)
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
