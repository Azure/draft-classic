package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/manifest"
)

const shipitDesc = `
Produce a packaged Docker build context and chart archive from the current directory.

These files are written to docker.tar.gz and chart.tar.gz, respectively. The intent
for 'shipit' is to

a) debug the docker build context and chart archive shipped to draftd
b) produce artifacts that can be further processed in a CI/CD system.
`

type shipitCmd struct {
	Client   *draft.Client
	Manifest *manifest.Manifest
	Out      io.Writer
	Src      string
}

func newShipitCmd(out io.Writer) *cobra.Command {
	var (
		shipit = &shipitCmd{
			Out: out,
		}
		runningEnvironment string
	)

	cmd := &cobra.Command{
		Use:   "shipit [path]",
		Short: "produce a docker build context and chart archive",
		Long:  shipitDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			shipit.Client = ensureDraftClient(shipit.Client)
			if len(args) > 0 {
				shipit.Src = args[0]
			}
			shipit.Client = ensureDraftClient(shipit.Client)
			shipit.Manifest = manifest.New()

			if shipit.Src == "" || shipit.Src == "." {
				shipit.Src, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			shipit.Manifest, err = loadDraftToml(shipit.Src)
			if err != nil {
				return err
			}

			return shipit.run(runningEnvironment)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&runningEnvironment, "environment", "e", defaultDraftEnvironment(), "the environment (development, staging, qa, etc) that draft will run under")

	return cmd
}

func (s *shipitCmd) run(environment string) error {
	if err := s.Client.ShipitFromDir(s.Manifest.Environments[environment].Name, s.Src); err != nil {
		return fmt.Errorf("Could not package a Docker build context and chart archive: %v", err)
	}
	return nil
}
