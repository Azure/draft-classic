// Copyright (c) Microsoft Corporation. All rights reserved.
//
// Licensed under the MIT license.

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

var globalUsage = "The draft server."

func newRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "draftd",
		Short:        "The draft server.",
		Long:         globalUsage,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flagDebug {
				log.Printf("debug logging enabled")
				log.SetLevel(log.DebugLevel)
			}
		},
	}
	p := cmd.PersistentFlags()
	p.BoolVar(&flagDebug, "debug", false, "enable verbose output")

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
