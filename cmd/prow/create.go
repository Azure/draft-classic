package main

import (
	"io"

	"github.com/spf13/cobra"
)

const createDesc = `This command transforms the local directory to be deployable via 'prow up'.
`

type createCmd struct {
	out io.Writer
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "transform the local directory to be deployable to Kubernetes",
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	return cmd
}

func (c *createCmd) run() error {
	return nil
}
