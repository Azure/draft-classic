package main

import (
	"io"

	"github.com/spf13/cobra"
)

var packManHelp = `The Draft Pack Manager.
`

func newPackManCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pack-man",
		Short: "manage Draft Packs",
		Long:  packManHelp,
	}
	cmd.AddCommand(
		newPackManAddCmd(out),
		newPackManListCmd(out),
		// newPackManRemoveCmd(out),
		newPackManUpdateCmd(out),
	)
	return cmd
}
