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
	appName   string
	client    prowd.Client
	namespace string
	out       io.Writer
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
	f.StringVarP(&up.appName, "app", "a", "", "name of helm release. By default this is the basename of the current working directory")
	f.StringVarP(&up.namespace, "namespace", "n", "default", "kubernetes namespace to install the chart")

	return cmd
}

func (u *upCmd) run() error {
	var err error
	if u.appName == "" {
		u.appName, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	if err = u.client.Up(u.appName, u.namespace, u.out); err != nil {
		return fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	return nil
}
