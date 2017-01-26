package main

import (
	"io"

	"github.com/spf13/cobra"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to the prow server.
`

type upCmd struct {
	out io.Writer
}

func newUpCmd(out io.Writer) *cobra.Command {
	cc := &upCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "up",
		Short: "upload the current directory to the prow server for deployment",
		Long:  upDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	return cmd
}

func (c *upCmd) run() error {
	return nil
}
