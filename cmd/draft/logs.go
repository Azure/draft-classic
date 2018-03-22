package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/spf13/cobra"
)

const logsDesc = `This command outputs logs from the draft server to help debug builds.`

type logsCmd struct {
	out      io.Writer
	appName  string
	buildID  string
	logLines int64
	args     []string
	home     draftpath.Home
}

func newLogsCmd(out io.Writer) *cobra.Command {
	lc := &logsCmd{
		out:  out,
		args: []string{"build-id"},
	}

	cmd := &cobra.Command{
		Use:     "logs <build-id>",
		Short:   logsDesc,
		Long:    logsDesc,
		PreRunE: lc.complete,
		RunE:    lc.run,
	}

	f := cmd.Flags()
	f.Int64Var(&lc.logLines, "tail", 100, "lines of recent log lines to display")

	return cmd
}

func (l *logsCmd) complete(_ *cobra.Command, args []string) error {
	if err := validateArgs(args, l.args); err != nil {
		return err
	}
	l.buildID = args[0]
	l.home = draftpath.Home(homePath())
	return nil
}

func (l *logsCmd) run(_ *cobra.Command, _ []string) error {
	f, err := os.Open(filepath.Join(l.home.Logs(), l.buildID))
	if err != nil {
		return fmt.Errorf("could not read logs for %s: %v", l.buildID, err)
	}
	defer f.Close()
	io.Copy(l.out, f)
	return nil
}
