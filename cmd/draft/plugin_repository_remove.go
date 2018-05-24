package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/spf13/cobra"
)

func newPluginRepositoryRemoveCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <rig...>",
		Short: "remove repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			home := plugin.Home(draftpath.Home(homePath()).Plugins())
			repositories := findRepositories(home.Repositories())
			foundRepositories := map[string]bool{}
			for _, arg := range args {
				foundRepositories[arg] = false
			}
			for _, repo := range repositories {
				for _, arg := range args {
					if repo == arg {
						foundRepositories[repo] = true
						if err := os.RemoveAll(filepath.Join(home.Repositories(), repo)); err != nil {
							return err
						}
					}
				}
			}
			t := time.Now()
			for repository, found := range foundRepositories {
				if !found {
					fmt.Fprintf(out, "repository '%s' was not found in the repository list\n", repository)
				}
			}
			fmt.Fprintf(out, "repositories uninstalled in %s\n", t.Sub(start).String())
			return nil
		},
	}
	return cmd
}
