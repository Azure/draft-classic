package main

import (
	"fmt"
	"io"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/gosuri/uitable"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newPluginSearchCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [keyword...]",
		Short: "perform a fuzzy search against available plugins",
		Run: func(cmd *cobra.Command, args []string) {
			home := plugin.Home(draftpath.Home(homePath()).Plugins())
			foundPlugins := search(args, home)
			table := uitable.New()
			table.AddRow("NAME", "REPOSITORY", "VERSION")
			for _, plugin := range foundPlugins {
				p, repository, err := getPlugin(plugin, home)
				if err == nil {
					table.AddRow(p.Name, repository, p.Version)
				} else {
					log.Debugln(err)
				}
			}
			fmt.Fprintln(out, table)
		},
	}
	return cmd
}
