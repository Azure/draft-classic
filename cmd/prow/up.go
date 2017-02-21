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
	appName      string
	client       *prow.Client
	namespace    string
	out          io.Writer
	buildTarPath string
	chartTarPath string
	wait         bool
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
	f.StringVar(&up.buildTarPath, "build-tar", "", "path to a gzipped build tarball. --chart-tar must also be set.")
	f.StringVar(&up.chartTarPath, "chart-tar", "", "path to a gzipped chart tarball. --build-tar must also be set.")
	f.BoolVarP(&up.wait, "wait", "w", false, "specifies whether or not to wait for all resources to be ready")

	return cmd
}

func (u *upCmd) run() (err error) {
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	if u.appName == "" {
		u.appName = path.Base(cwd)
	}
	u.client.OptionWait = u.wait
	if u.buildTarPath != "" && u.chartTarPath != "" {
		buildTar, e := os.Open(u.buildTarPath)
		if e != nil {
			return e
		}
		chartTar, e := os.Open(u.chartTarPath)
		if e != nil {
			return e
		}
		err = u.client.Up(u.appName, u.namespace, u.out, buildTar, chartTar)
	} else {
		err = u.client.UpFromDir(u.appName, u.namespace, u.out, cwd)
	}

	// format error before returning
	if err != nil {
		err = fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	return
}
