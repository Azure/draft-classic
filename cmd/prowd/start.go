package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/helm/prow/api"
)

const startDesc = `
Starts the prow server.
`

type startCmd struct {
	out io.Writer
	// listenAddr is the address which the server will be listening on.
	listenAddr string
}

func newStartCmd(out io.Writer) *cobra.Command {
	sc := &startCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the prow server",
		Long:  startDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sc.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.listenAddr, "listen-addr", "l", "tcp://0.0.0.0:8080", "the address the server listens on")

	return cmd
}

func (c *startCmd) run() error {
	protoAndAddr := strings.SplitN(c.listenAddr, "://", 2)
	server, err := api.NewServer(protoAndAddr[0], protoAndAddr[1])
	if err != nil {
		return fmt.Errorf("failed to create server at %s: %v", c.listenAddr, err)
	}
	log.Printf("server is now listening at %s", c.listenAddr)
	if err = server.Serve(); err != nil {
		return err
	}
	return nil
}
