package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type configSetCmd struct {
	out   io.Writer
	args  []string
	key   string
	value string
}

func newConfigSetCmd(out io.Writer) *cobra.Command {
	ccmd := &configSetCmd{
		out:  out,
		args: []string{"key", "value"},
	}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set global Draft configuration stored in $DRAFT_HOME/config.toml",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ccmd.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ccmd.run()
		},
	}
	return cmd
}

func (ccmd *configSetCmd) complete(args []string) error {
	if err := validateConfigSetArgs(args, ccmd.args); err != nil {
		return err
	}
	ccmd.key = args[0]
	ccmd.value = args[1]
	return nil
}

func (ccmd *configSetCmd) run() error {
	if globalConfig == nil {
		return fmt.Errorf("Draft configuration in $DRAFT_HOME/config.toml has not been initialized. Run draft init to get started.")
	}
	globalConfig[ccmd.key] = ccmd.value
	return SaveConfig(globalConfig)
}

func validateConfigSetArgs(args, expectedArgs []string) error {
	if len(args) == 1 {
		for _, k := range configKeys {
			if k.name == args[0] {
				return fmt.Errorf("This command needs a value: %v", k.description)
			}
		}
		return fmt.Errorf("This command needs a value. No help available for key '%v'", args[0])
	}
	if len(args) != len(expectedArgs) {
		keys := []string{}
		for _, k := range configKeys {
			keys = append(keys, "  "+k.name+": "+k.description)
		}
		return fmt.Errorf("This command needs a key and a value to set it to. Supported keys:\n%v", strings.Join(keys, "\n"))
	}
	return nil
}
