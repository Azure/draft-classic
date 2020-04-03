package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

//TODO: make sure this description looks pretty on command line
const packCreateDesc = `
This command creates a pack directory along with the common files and
directories used in a Draft pack

For example, 'draft pack create foo' will create a directory structure that looks
something like this:

	foo/
	  |
	  |- Dockerfile     # Sample Dockerfile for your application
	  |
	  |- charts/        # Contains a sample chart for your application
	  |
	  |- .dockerignore  # Contains files to exclude when building Docker image
	  |
	  |- tasks.toml     # Lets you specify tasks to run during development

'draft pack create' takes a name or a path for an argument. If a path is given, the last element in the path will be used as the name for the pack. If directories in the given path do not exist, Draft will attempt to create them as it goes. If the given destination exists and there are files in that directory, draft will return an error.
`

type packCreateCmd struct {
	out  io.Writer
	name string
}

func newPackCreateCmd(out io.Writer) *cobra.Command {
	cc := &packCreateCmd{out: out}
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "scaffold a new draft pack with a given name",
		Long:  packCreateDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("a name for the draft pack is required")
			} else if len(args) > 1 {
				return errors.New("this command takes one argument: name")
			}
			cc.name = args[0]

		},
	}
	return cmd
}

func (cmd *packCreateCmd) run() error {
	packs, err := pack.Create(cmd.name)
	if err != nil {
		return err
	}
	//TODO: print paths to files that were created like rails new
	fmt.Fprintf(cmd.out, "Created draft pack %s\nEdit with your application's details and configure to your heart's content!\n", cmd.name)
	return nil
}
