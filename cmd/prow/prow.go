package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug       bool
)

var globalUsage = `The application deployment tool for Kubernetes.

Commands:

- prow create:    transform the local directory to be deployable via prow
- prow push:      packages and deploys your local directory to Kubernetes
- prow version:   display client version information
`

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "prow",
		Short:        "The application deployment tool for Kubernetes.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	p := cmd.PersistentFlags()
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")
	
	cmd.AddCommand(
		newCreateCmd(out),
		newPushCmd(out),
		newVersionCmd(out),
	)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
