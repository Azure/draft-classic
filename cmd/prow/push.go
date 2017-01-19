package main

import (
	"io"

	"github.com/spf13/cobra"
)

const pushDesc = `
This command archives the current directory into a tar archive and uploads it to the prow server.
`

type pushCmd struct {
	out     io.Writer
}

func newPushCmd(out io.Writer) *cobra.Command {
	cc := &pushCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "push",
		Short: "upload the current directory to the prow server for deployment",
		Long:  pushDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	return cmd
}

func (c *pushCmd) run() error {
	return nil
}
