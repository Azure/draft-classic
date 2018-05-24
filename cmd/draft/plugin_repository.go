package main

import (
	"io"

	"github.com/spf13/cobra"
)

const (
	pluginRepoHelp = `Manage Draft plugin repositories.`
)

func newPluginRepositoryCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repository",
		Short: "manage Draft plugin repositories",
		Long:  pluginRepoHelp,
	}
	cmd.AddCommand(
		newPluginRepositoryAddCmd(out),
		newPluginRepositoryListCmd(out),
		newPluginRepositoryRemoveCmd(out),
	)
	return cmd
}
