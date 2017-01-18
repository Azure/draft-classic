package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/helm/prow/cmd/prow/prowpath"
)

const initDesc = `
This command will bootstrap your prow config directory with a helm chart skeleton that will be used
with future prow commands. It will also set up any other necessary local configuration.
`

type initCmd struct {
	home    prowpath.Home
	out     io.Writer
	starter string
}

func newInitCmd(out io.Writer) *cobra.Command {
	cc := &initCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize prow client configuration",
		Long:  initDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			cc.home = prowpath.Home(homePath())
			return cc.run()
		},
	}

	return cmd
}

func (c *initCmd) run() error {
	return nil
}
