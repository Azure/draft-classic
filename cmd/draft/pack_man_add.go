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

type packManAddCmd struct {
	out    io.Writer
	err    io.Writer
	source string
	home   draftpath.Home
}

func newPackManAddCmd(out io.Writer) *cobra.Command {
	add := &packManAddCmd{out: out}

	cmd := &cobra.Command{
		Use:   "add [flags] <path|url>",
		Short: "Add a Pack repository",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return add.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return add.run()
		},
	}
	return cmd
}

func (a *packManAddCmd) complete(args []string) error {
	if err := validateArgs(args, []string{"path|url"}); err != nil {
		return err
	}
	a.source = args[0]
	a.home = draftpath.Home(homePath())
	return nil
}

func (a *packManAddCmd) run() error {
	i, err := installer.New(a.source, "", a.home)
	if err != nil {
		return err
	}

	debug("installing pack repo from %s", a.source)
	if err := installer.Install(i); err != nil {
		return err
	}

	p := repo.Repository{
		Name: filepath.Base(i.Path()),
		Dir:  filepath.Dir(i.Path()),
	}

	fmt.Fprintf(a.out, "Installed pack repository %s\n", p.Name)
	return nil
}
