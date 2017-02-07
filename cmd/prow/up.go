package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/deis/prow/pkg/prowd"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to the prow server.
`

type upCmd struct {
	out       io.Writer
	client    prowd.Client
	namespace string
}

func newUpCmd(out io.Writer) *cobra.Command {
	up := &upCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:     "up",
		Short:   "upload the current directory to the prow server for deployment",
		Long:    upDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			up.client = ensureProwClient(up.client)
			return up.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&up.namespace, "namespace", "default", "kubernetes namespace to install the chart")

	return cmd
}

func (u *upCmd) run() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := u.client.Up(currentDir, u.namespace); err != nil {
		return fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	return nil
}
