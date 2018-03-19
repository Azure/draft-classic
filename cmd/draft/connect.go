package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/local"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const (
	connectDesc = `This command creates a local environment for you to test your app. It will give you a localhost urls that you can use to see your application working and it will print out logs from your application. This command must be run in the root of your application directory.
`
)

var (
	targetContainer string
	overridePorts   []string

	waitFlag bool
)

type connectCmd struct {
	client   *draft.Client
	out      io.Writer
	logLines int64
}

func newConnectCmd(out io.Writer) *cobra.Command {
	var (
		cc                 = &connectCmd{out: out}
		runningEnvironment string
	)

	cmd := &cobra.Command{
		Use:     "connect",
		Short:   "connect to your application locally",
		Long:    connectDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			cc.client = ensureDraftClient(cc.client)
			return cc.run(runningEnvironment)
		},
	}

	f := cmd.Flags()
	f.Int64Var(&cc.logLines, "tail", 5, "lines of recent log lines to display")
	f.StringVarP(&runningEnvironment, environmentFlagName, environmentFlagShorthand, defaultDraftEnvironment(), environmentFlagUsage)
	f.StringSliceVarP(&overridePorts, "override-port", "p", []string{}, "specify a local port to connect to, in the form <local>:<remote>")
	f.StringVarP(&targetContainer, "container", "c", "", "name of the container to connect to")
	f.BoolVarP(&waitFlag, "wait", "w", false, "exits with code 0 when draft can connect")

	return cmd
}

func (cn *connectCmd) run(runningEnvironment string) (err error) {
	deployedApp, err := local.DeployedApplication(draftToml, runningEnvironment)
	if err != nil {
		return err
	}

	client, config, err := getKubeClient(kubeContext)
	if err != nil {
		return err
	}

	var ports []string
	if len(overridePorts) == 0 {
		if deployedApp.OverridePorts != "" {
			// removes multiple spaces
			s := strings.Join(strings.Fields(deployedApp.OverridePorts), " ")
			ports = strings.Split(s, " ")
		}
	} else {
		ports = overridePorts
	}

	buildID, err := cn.client.GetLatestBuildID(context.Background(), deployedApp.Name)
	if err != nil {
		return fmt.Errorf("cannot get latest build id: %v", err)
	}

	fmt.Fprintf(cn.out, "Connecting to your application...\n")
	connection, err := deployedApp.Connect(client, config, buildID, targetContainer, ports)
	if err != nil {
		return err
	}

	if waitFlag {
		return
	}

	var connectionMessage = "Your connection is still active. \n"

	for _, cc := range connection.ContainerConnections {
		for _, t := range cc.Tunnels {
			err = t.ForwardPort()
			if err != nil {
				return err
			}
			m := fmt.Sprintf("Connect to %v:%v on localhost:%#v\n", cc.ContainerName, t.Remote, t.Local)
			connectionMessage += m
			fmt.Fprintf(cn.out, m)
		}
	}

	for _, cc := range connection.ContainerConnections {
		readCloser, err := connection.RequestLogStream(deployedApp.Namespace, cc.ContainerName, cn.logLines)
		if err != nil {
			return err
		}

		defer readCloser.Close()
		go writeContainerLogs(cn.out, readCloser, cc.ContainerName)
	}

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		os.Exit(0)
	}()

	for {
		fmt.Fprintf(cn.out, connectionMessage)
		time.Sleep(10 * time.Second)
	}
}

func writeContainerLogs(out io.Writer, in io.ReadCloser, containerName string) error {
	b := bufio.NewReader(in)
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "[%v]: %v", containerName, line)
	}
}
