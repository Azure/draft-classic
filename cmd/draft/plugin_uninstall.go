package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"

	"github.com/spf13/cobra"
)

type pluginUninstallCmd struct {
	names []string
	home  draftpath.Home
	out   io.Writer
}

func newPluginUninstallCmd(out io.Writer) *cobra.Command {
	pcmd := &pluginUninstallCmd{out: out}
	cmd := &cobra.Command{
		Use:   "uninstall <plugin>...",
		Short: "uninstall one or more Draft plugins",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.run()
		},
	}
	return cmd
}

func (pcmd *pluginUninstallCmd) complete(args []string) error {
	if len(args) == 0 {
		return errors.New("please provide plugin name to remove")
	}
	pcmd.names = args
	pcmd.home = draftpath.Home(homePath())
	return nil
}

func (pcmd *pluginUninstallCmd) run() error {
	pHome := plugin.Home(pcmd.home.Plugins())
	for _, plugin := range pcmd.names {
		relevantPlugins := search([]string{plugin}, pHome)
		switch len(relevantPlugins) {
		case 0:
			return fmt.Errorf("no plugin with the name '%s' was found", plugin)
		case 1:
			plugin = relevantPlugins[0]
		default:
			var match bool
			// check if we have an exact match
			for _, f := range relevantPlugins {
				if strings.Compare(f, plugin) == 0 {
					plugin = f
					match = true
				}
			}
			if !match {
				return fmt.Errorf("%d plugins with the name '%s' was found: %v", len(relevantPlugins), plugin, relevantPlugins)
			}
		}
		plugin, _, err := getPlugin(plugin, pHome)
		if err != nil {
			return err
		}
		fmt.Fprintf(pcmd.out, "Uninstalling %s...\n", plugin.Name)
		start := time.Now()
		if err := plugin.Uninstall(pHome); err != nil {
			return err
		}
		t := time.Now()
		fmt.Fprintf(pcmd.out, "%s %s: uninstalled in %s\n", plugin.Name, plugin.Version, t.Sub(start).String())
		return nil
	}
	return nil
}
