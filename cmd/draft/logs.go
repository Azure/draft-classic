package main

import (
	"fmt"
	"io"
	"github.com/Azure/draft/pkg/draft"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const logsDesc = `This command outputs logs from the draft server to help debug builds.`

type logsCmd struct {
	client   *draft.Client
	out      io.Writer
	buildID  string
	logLines int64
}

func newLogsCmd(out io.Writer) *cobra.Command {
	lc := &logsCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:     "logs",
		Short:   logsDesc,
		Long:    logsDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing build id")
			}
			lc.buildID = args[0]
			lc.client = ensureDraftClient(lc.client)
			return lc.run()
		},
	}

	f := cmd.Flags()
	f.Int64Var(&lc.logLines, "tail", 100, "lines of recent log lines to display")

	return cmd
}

func (l *logsCmd) run() error {
	b, err := l.client.GetLogs(context.Background(), l.buildID, draft.WithLogsLimit(l.logLines))
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}
