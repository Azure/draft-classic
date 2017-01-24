package main

import (
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug bool
)

var globalUsage = "The prow server."

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "prowd",
		Short:        "The prow server.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	p := cmd.PersistentFlags()
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")

	if flagDebug {
		log.SetLevel(log.DebugLevel)
	}

	cmd.AddCommand(
		newStartCmd(out),
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
