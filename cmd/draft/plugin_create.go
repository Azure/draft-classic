package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/plugin"
	"github.com/Azure/draft/pkg/plugin/repository"

	"github.com/spf13/cobra"
)

const createTpl = `local name = "{{ .Name }}"
local version = "0.1.0"

food = {
    name = name,
    description = "enter description here",
    homepage = "https://github.com/Azure/draft",
	version = version,
	useTunnel = false,
    packages = {
        {
            os = "darwin",
            arch = "amd64",
            url = "",
            sha256 = "",
        },
        {
            os = "linux",
            arch = "amd64",
            url = "",
            sha256 = "",
        },
        {
            os = "windows",
            arch = "amd64",
            url = "",
            sha256 = "",
        }
    }
}
`

func newPluginCreateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "generate plugin scripts and display the file path on stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := plugin.Home(draftpath.Home(homePath()).Plugins())
			destPath := filepath.Join(home.Repositories(), repository.DefaultPluginRepository, "Plugins", fmt.Sprintf("%s.lua", args[0]))
			f, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t := template.Must(template.New("create").Parse(createTpl))
			if err := t.Execute(f, struct{ Name string }{args[0]}); err != nil {
				return err
			}
			fmt.Fprintln(out, destPath)
			return nil
		},
	}
	return cmd
}
