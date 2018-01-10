package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/local"
)

const (
	connectDesc = `This command creates a local environment for you to test your app. It will give you a localhost url that you can use to see your application working and it will print out logs from your application. This command must be run in the root of your application directory.
`
)

type connectCmd struct {
	out      io.Writer
	logLines int64
}

func newConnectCmd(out io.Writer) *cobra.Command {
	var (
		cc          = &connectCmd{out: out}
		environment string
	)

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "connect to your application locally",
		Long:  connectDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(environment)
		},
	}

	f := cmd.Flags()
	f.Int64Var(&cc.logLines, "tail", 5, "lines of recent log lines to display")
	f.StringVarP(&environment, "environment", "e", defaultDraftEnvironment(), "the environment (development, staging, qa, etc) that draft will run under")

	return cmd
}

func (cn *connectCmd) run(environment string) (err error) {
	deployedApp, err := local.DeployedApplication(draftToml, environment)
	if err != nil {
		return err
	}

	client, config, err := getKubeClient(kubeContext)
	if err != nil {
		return err
	}

	fmt.Fprintf(cn.out, "Connecting to your app...")
	connection, err := deployedApp.Connect(client, config)
	if err != nil {
		return err
	}

	detail := fmt.Sprintf("localhost:%#v", connection.Tunnel.Local)
	fmt.Fprintln(cn.out, "SUCCESS...Connect to your app on "+detail)

	fmt.Fprintln(cn.out, "Starting log streaming...")
	readCloser, err := connection.RequestLogStream(deployedApp, cn.logLines)
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(cn.out, readCloser)
	if err != nil {
		return err
	}

	return nil
}
