package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/pack/repo"
	"github.com/Azure/draft/pkg/draft/pack/repo/installer"
)

type packManUpdateCmd struct {
	out    io.Writer
	err    io.Writer
	source string
	home   draftpath.Home
}

func newPackManUpdateCmd(out io.Writer) *cobra.Command {
	upd := &packManUpdateCmd{out: out}

	cmd := &cobra.Command{
		Use:   "update [flags]",
		Short: "Fetch the newest version of all packs using git.",
		RunE: func(cmd *cobra.Command, args []string) error {
			upd.home = draftpath.Home(homePath())
			return upd.run()
		},
	}
	return cmd
}

func (upd *packManUpdateCmd) run() error {

	repos := repo.FindRepositories(upd.home.Packs())
	if len(repos) == 0 {
		fmt.Fprintf(upd.out, "No pack repositories found to update. All up to date!")
		return nil
	}
	var updatedRepoNames []string
	for i := range repos {
		exactLocation, err := filepath.EvalSymlinks(repos[i].Dir)
		if err != nil {
			return err
		}
		absExactLocation, err := filepath.Abs(exactLocation)
		if err != nil {
			return err
		}

		i, err := installer.FindSource(absExactLocation, upd.home)
		if err != nil {
			return err
		}
		if err := installer.Update(i); err != nil {
			return err
		}
		updatedRepoNames = append(updatedRepoNames, filepath.Base(i.Path()))
	}
	repoMsg := "Updated %d pack repository %v\n"
	if len(repos) > 1 {
		repoMsg = "Updated %d pack repositories %v\n"
	}
	fmt.Fprintf(upd.out, repoMsg, len(repos), updatedRepoNames)
	return nil
}
