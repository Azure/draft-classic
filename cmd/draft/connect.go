package main

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/local"
)

const (
	connectDesc = `This command creates a local environment for you to test your app. It will give you a localhost url that you can use to see your application working and it will print out logs from your application. This command must be run in the root of your application directory.
`
)

type connectCmd struct {
	out io.Writer
}

func newConnectCmd(out io.Writer) *cobra.Command {
	cc := &connectCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "connect to your application locally",
		Long:  connectDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	return cmd
}

func (cn *connectCmd) run() (err error) {

	deployedApp, err := local.DeployedApplication(draftToml, defaultDraftEnvironment())
	if err != nil {
		return err
	}

	clientset, config, err := getKubeClient(kubeContext)
	if err != nil {
		return err
	}
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return err
	}

	fmt.Fprintf(cn.out, "Connecting to your app...")
	connection, err := deployedApp.Connect(clientset, clientConfig)
	if err != nil {
		return err
	}

	detail := fmt.Sprintf("localhost:%#v", connection.Tunnel.Local)
	fmt.Fprintln(cn.out, "SUCCESS...Connect to your app on "+detail)

	fmt.Fprintln(cn.out, "Starting log streaming...")
	readCloser, err := connection.RequestLogStream(deployedApp)
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(cn.out, readCloser)
	if err != nil {
		return err
	}

	time.Sleep(40 * time.Minute) //TODO: put this in wait loop and let the logs roll

	return nil
}
