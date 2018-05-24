package main

import (
	"fmt"
	"io"
	"time"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/spf13/cobra"
)

type pluginUpgradeCmd struct {
	home draftpath.Home
	out  io.Writer
}

func newPluginUpgradeCmd(out io.Writer) *cobra.Command {
	pcmd := &pluginUpgradeCmd{out: out}
	cmd := &cobra.Command{
		Use:   "upgrade [plugin..]",
		Short: "upgrade all plugins. If arguments are provided, only the specified plugins are upgraded.",
		RunE: func(cmd *cobra.Command, args []string) error {
			pcmd.home = draftpath.Home(homePath())
			return pcmd.run(args)
		},
	}
	return cmd
}

func (pcmd *pluginUpgradeCmd) run(args []string) error {
	pHome := plugin.Home(pcmd.home.Plugins())
	if err := updatePluginRepositories(pHome); err != nil {
		return err
	}
	var pluginNames []string
	if len(args) > 0 {
		pluginNames = args
	} else {
		pluginNames = findPlugins(pHome)
	}
	for _, name := range pluginNames {
		plugin, _, err := getPlugin(name, pHome)
		if err != nil {
			return err
		}
		fmt.Fprintf(pcmd.out, "Upgrading %s...\n", plugin.Name)
		start := time.Now()
		if err := plugin.Install(pHome); err != nil {
			return err
		}
		t := time.Now()
		fmt.Fprintf(pcmd.out, "%s %s: upgraded in %s\n", plugin.Name, plugin.Version, t.Sub(start).String())
	}
	return nil
}
