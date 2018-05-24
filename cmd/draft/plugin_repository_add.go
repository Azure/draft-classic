package main

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/Azure/draft/pkg/plugin/repository/installer"
)

func newPluginRepositoryAddCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <repository>",
		Short: "add repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := plugin.Home(draftpath.Home(homePath()).Plugins())
			i, err := installer.New(args[0], "", home)
			if err != nil {
				return err
			}

			start := time.Now()
			if err := installer.Install(i); err != nil {
				return err
			}
			t := time.Now()
			fmt.Fprintf(out, "repository added in %s\n", t.Sub(start).String())
			return nil
		},
	}
	return cmd
}
