package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/helm/prow/cmd/prow/prowpath"
)

const pushDesc = `
This command archives the current directory into a tarball and uploads it to
prowd.
`

type pushCmd struct {
	home    prowpath.Home
	out     io.Writer
	starter string
}

func newPushCmd(out io.Writer) *cobra.Command {
	cc := &pushCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "push",
		Short: "upload the current directory to prowd for deployment",
		Long:  pushDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			cc.home = prowpath.Home(homePath())
			return cc.run()
		},
	}

	return cmd
}

func (c *pushCmd) run() error {
	return nil
}
