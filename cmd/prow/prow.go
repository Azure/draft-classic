package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	prowHome        string
	// flagDebug is a signal that the user wants additional output.
	flagDebug       bool
)

var globalUsage = `The application deployment tool for Kubernetes.

To begin working with prow, run the 'prow init' command:

	$ prow init

This will bootstrap your prow config directory with a helm chart skeleton that will be used with
future prow commands. It will also set up any other necessary local configuration.

Common actions from this point include:

- prow create:    transform the local directory to be deployable via prow
- prow push:      packages and deploys your local directory to Kubernetes
- prow list:      list applications deployed to Kubernetes
`

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "prow",
		Short:        "The application deployment tool for Kubernetes.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	p := cmd.PersistentFlags()
	p.StringVar(&prowHome, "home", defaultProwHome(), "location of your Prow config.")
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	
	cmd.AddCommand(
		newCreateCmd(out),
		newHomeCmd(out),
		newInitCmd(out),
		newPushCmd(out),
		newVersionCmd(out),
	)

	return cmd
}

func defaultProwHome() string {
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
