package main

import (
	"fmt"
	"io"

	"github.com/Azure/draft/pkg/draft/draftpath"

	"github.com/Azure/draft/pkg/plugin"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

type pluginListCmd struct {
	home draftpath.Home
	out  io.Writer
}

func newPluginListCmd(out io.Writer) *cobra.Command {
	pcmd := &pluginListCmd{out: out}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list installed plugins. If an argument is provided, list all installed versions of that plugin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pcmd.home = draftpath.Home(homePath())
			return pcmd.run(args)
		},
	}
	return cmd
}

func (pcmd *pluginListCmd) run(args []string) error {
	table := uitable.New()
	pHome := plugin.Home(pcmd.home.Plugins())
	if len(args) == 0 {
		table.AddRow("NAME")
		for _, plugin := range findPlugins(pHome) {
			table.AddRow(plugin)
		}
	} else {
		table.AddRow("NAME", "VERSION")
		for _, ver := range findPluginVersions(args[0], pHome) {
			p := plugin.Plugin{
				Name:    args[0],
				Version: ver,
			}
			table.AddRow(p.Name, p.Version)
		}
	}
	fmt.Println(table)
	return nil
}
