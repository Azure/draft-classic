package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
)

type pluginInstallCmd struct {
	name    string
	version string
	home    draftpath.Home
	out     io.Writer
	args    []string
}

func newPluginInstallCmd(out io.Writer) *cobra.Command {
	pcmd := &pluginInstallCmd{
		out:  out,
		args: []string{"plugin"},
	}

	cmd := &cobra.Command{
		Use:   "install [options] <name>",
		Short: "install Draft plugins",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcmd.run()
		},
	}
	return cmd
}

func (pcmd *pluginInstallCmd) complete(args []string) error {
	if err := validateArgs(args, pcmd.args); err != nil {
		return err
	}
	pcmd.name = args[0]
	pcmd.home = draftpath.Home(homePath())
	return nil
}

func (pcmd *pluginInstallCmd) run() error {
	pHome := plugin.Home(pcmd.home.Plugins())
	relevantPlugins := search([]string{pcmd.name}, pHome)
	switch len(relevantPlugins) {
	case 0:
		return fmt.Errorf("no plugins with the name '%s' was found", pcmd.name)
	case 1:
		pcmd.name = relevantPlugins[0]
	default:
		var match bool
		// check if we have an exact match
		for _, f := range relevantPlugins {
			if strings.Compare(f, pcmd.name) == 0 {
				pcmd.name = f
				match = true
			}
		}
		if !match {
			return fmt.Errorf("%d plugins with the name '%s' was found: %v", len(relevantPlugins), pcmd.name, relevantPlugins)
		}
	}
	plug, _, err := getPlugin(pcmd.name, pHome)
	if err != nil {
		return err
	}
	if len(findPluginVersions(pcmd.name, pHome)) > 0 {
		fmt.Fprintf(pcmd.out, "%s is already installed. Please use `draft plugin upgrade %s` to upgrade.\n", pcmd.name, pcmd.name)
		return nil
	}
	fmt.Fprintf(pcmd.out, "Installing %s...\n", pcmd.name)
	start := time.Now()
	if err := plug.Install(pHome); err != nil {
		return err
	}
	t := time.Now()
	fmt.Fprintf(pcmd.out, "%s %s: installed in %s\n", plug.Name, plug.Version, t.Sub(start).String())
	return nil
}
