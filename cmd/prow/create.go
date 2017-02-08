package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/deis/prow/pkg/dockerutil"
	"github.com/deis/prow/pkg/prow/prowpath"
)

const (
	createDesc = `This command transforms the local directory to be deployable via 'prow up'.
`
	defaultDockerfile = `FROM nginx:latest
`
)

type createCmd struct {
	appName string
	out     io.Writer
	pack    string
	home    prowpath.Home
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "transform the local directory to be deployable to Kubernetes",
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	cc.home = prowpath.Home(homePath())

	f := cmd.Flags()
	f.StringVarP(&cc.appName, "app", "a", "", "name of the Helm release. By default this is the basename of the current working directory")
	f.StringVarP(&cc.pack, "pack", "p", "", "the named Prow starter pack to scaffold the app with")

	return cmd
}

func (c *createCmd) run() error {
	var err error

	if c.appName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		c.appName = path.Base(cwd)
	}

	cfile := &chart.Metadata{
		// HACK(bacongobbler): chartutil.Create uses the name as the directory. Because we want to
		// write it to chart/, we name the chart chart and then re-save the Chart.yaml later with
		// chartutil.SaveChartfile().
		Name:        "chart",
		Description: "A Helm chart for Kubernetes",
		Version:     "0.1.0",
		ApiVersion:  chartutil.ApiVersionV1,
	}

	if c.pack != "" {
		// Create a chart from the starter pack
		lpack := filepath.Join(c.home.Packs(), c.pack)
		err = chartutil.CreateFrom(cfile, "", lpack)
	} else {
		_, err = chartutil.Create(cfile, "")
	}

	if err != nil {
		if os.IsExist(err) {
			// chart dir already exists, so we just tell the user that we are happily skipping
			fmt.Fprintln(c.out, "--> Chart already exists at chart/, skipping")
		} else {
			return fmt.Errorf("there was an error creating the chart: %v", err)
		}
	} else {
		// HACK(bacongobbler): see comment above about chartutil.Create
		cfile.Name = c.appName
		if err := chartutil.SaveChartfile(path.Join("chart", "Chart.yaml"), cfile); err != nil {
			return fmt.Errorf("there was an error creating the chart: %v", err)
		}
	}

	// now we check for a Dockerfile and create that based on the starter pack
	if c.pack != "" {
		lpack := filepath.Join(c.home.Packs(), c.pack, "Dockerfile")
		err = dockerutil.CreateFrom("Dockerfile", lpack)
	} else {
		err = dockerutil.Create("Dockerfile", bytes.NewBufferString(defaultDockerfile))
	}

	if err != nil {
		if os.IsExist(err) {
			// Dockerfile already exists, so we just tell the user that we are happily skipping
			fmt.Fprintln(c.out, "--> Dockerfile already exists, skipping")
		} else {
			return fmt.Errorf("there was an error creating the Dockerfile: %v", err)
		}
	}

	fmt.Fprintln(c.out, "--> Ready to sail")
	return nil
}
