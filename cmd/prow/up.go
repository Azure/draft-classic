package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/deis/prow/pkg/prow"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to the prow server.
`

type upCmd struct {
	appName   string
	client    *prow.Client
	namespace string
	out       io.Writer
	wait      bool
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
	f.StringVarP(&up.appName, "app", "a", "", "name of the helm release. By default this is the basename of the current working directory")
	f.StringVarP(&up.namespace, "namespace", "n", "default", "kubernetes namespace to install the chart")
	f.BoolVarP(&up.wait, "wait", "w", false, "specifies whether or not to wait for all resources to be ready")

	return cmd
}

func (u *upCmd) run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if u.appName == "" {
		u.appName = path.Base(cwd)
	}
	u.client.OptionWait = u.wait
	if err := u.client.Up(u.appName, cwd, u.namespace, u.out); err != nil {
		return fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	return nil
}
