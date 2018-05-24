package main

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
)

type pluginUpdateCmd struct {
	names []string
	home  draftpath.Home
	out   io.Writer
}

func newPluginUpdateCmd(out io.Writer) *cobra.Command {
	pcmd := &pluginUpdateCmd{out: out}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update plugin repositories",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.run()
		},
	}
	return cmd
}

func (pcmd *pluginUpdateCmd) complete(args []string) error {
	pcmd.home = draftpath.Home(homePath())
	return nil
}

func (pcmd *pluginUpdateCmd) run() error {
	start := time.Now()
	home := plugin.Home(pcmd.home.Plugins())
	if err := updatePluginRepositories(home); err != nil {
		return err
	}
	t := time.Now()
	fmt.Fprintf(pcmd.out, "repositories updated in %s\n", t.Sub(start).String())
	return nil
}
