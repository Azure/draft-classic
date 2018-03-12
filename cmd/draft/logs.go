package main

import (
	"fmt"
	"io"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/local"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const logsDesc = `This command outputs logs from the draft server to help debug builds.`

type logsCmd struct {
	client             *draft.Client
	out                io.Writer
	appName            string
	buildID            string
	logLines           int64
	args               []string
	runningEnvironment string
}

func newLogsCmd(out io.Writer) *cobra.Command {

	lc := &logsCmd{
		out:  out,
		args: []string{"build-id"},
	}

	cmd := &cobra.Command{
		Use:   "logs <build-id>",
		Short: logsDesc,
		Long:  logsDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setupConnection(cmd, args); err != nil {
				return err
			}
			return lc.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			lc.client = ensureDraftClient(lc.client)
			return lc.run()
		},
	}

	f := cmd.Flags()
	f.Int64Var(&lc.logLines, "tail", 100, "lines of recent log lines to display")
	f.StringVarP(&lc.runningEnvironment, environmentFlagName, environmentFlagShorthand, defaultDraftEnvironment(), environmentFlagUsage)

	return cmd
}

func (l *logsCmd) complete(args []string) error {
	if err := validateArgs(args, l.args); err != nil {
		return err
	}
	l.buildID = args[0]
	return nil
}

func (l *logsCmd) run() error {
	deployedApp, err := local.DeployedApplication(draftToml, l.runningEnvironment)
	if err != nil {
		return err
	}

	b, err := l.client.GetLogs(context.Background(), deployedApp.Name, l.buildID, draft.WithLogsLimit(l.logLines))
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}
