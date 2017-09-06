package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/pack/repo"
)

type packManListCmd struct {
	out  io.Writer
	home draftpath.Home
}

func newPackManListCmd(out io.Writer) *cobra.Command {
	list := &packManListCmd{out: out}

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List all installed pack repositories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			list.home = draftpath.Home(homePath())
			return list.run()
		},
	}
	return cmd
}

func (l *packManListCmd) run() error {
	repos := repo.FindRepositories(l.home.Packs())
	if len(repos) == 0 {
		return errors.New("no pack repositories to show")
	}
	table := uitable.New()
	table.AddRow("NAME", "PATH")
	for i := range repos {
		table.AddRow(repos[i].Name, repos[i].Dir)
	}
	fmt.Fprintln(l.out, table)
	return nil
}
