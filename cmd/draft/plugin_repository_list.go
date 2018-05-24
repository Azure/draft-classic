package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func newPluginRepositoryListCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := plugin.Home(draftpath.Home(homePath()).Plugins())
			repositories := findRepositories(home.Repositories())
			table := uitable.New()
			table.AddRow("NAME")
			for _, repo := range repositories {
				table.AddRow(repo)
			}
			fmt.Fprintln(out, table)
			return nil
		},
	}
	return cmd
}

func findRepositories(dir string) []string {
	var repositories []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && f.Name() == "Plugins" {
			repoName := filepath.ToSlash(strings.TrimPrefix(filepath.Dir(path), dir+string(os.PathSeparator)))
			repositories = append(repositories, repoName)
		}
		return nil
	})
	return repositories
}
