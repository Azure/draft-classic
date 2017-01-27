package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deis/prow/pkg/prowd"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to the prow server.
`

type upCmd struct {
	out    io.Writer
	client prowd.Client
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
	return cmd
}

func (u *upCmd) run() error {
	release, err := u.client.Up(".", "default")
	if err != nil {
		return fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	fmt.Println(release)
	return nil
}
