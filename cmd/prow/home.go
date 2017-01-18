package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var longHomeHelp = `
This command displays the location of prow's home directory. This is where any prow configuration
files live.
`

func newHomeCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "home",
		Short: "displays the location of prow's home directory",
		Long:  longHomeHelp,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(out, "%s\n", homePath())
		},
	}

	return cmd
}
