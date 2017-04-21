package main

import (
	"errors"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/deis/draft/pkg/draft"
	"github.com/deis/draft/pkg/version"
)

const versionDesc = `
Show the client version for draft.

This will print the client version of draft. The output will look something like
this:

Client: &version.Version{SemVer:"v0.1.0", GitCommit:"4f97233d2cc2c7017b07f94211e55bb2670f990d", GitTreeState:"clean"}
Server: &version.Version{SemVer:"v0.1.0", GitCommit:"4f97233d2cc2c7017b07f94211e55bb2670f990d", GitTreeState:"clean"}
`

type versionCmd struct {
	out        io.Writer
	client     *draft.Client
	showClient bool
	showServer bool
	short      bool
}

func newVersionCmd(out io.Writer) *cobra.Command {
	version := &versionCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the client version information",
		Long:  versionDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If neither is explicitly set, show both.
			if !version.showClient && !version.showServer {
				version.showClient, version.showServer = true, true
			}
			if version.showServer {
				// We do this manually instead of in PreRun because we only
				// need a tunnel if server version is requested.
				setupConnection(cmd, args)
			}
			version.client = ensureDraftClient(version.client)
			return version.run()
		},
	}
	return cmd
}

func (v *versionCmd) run() error {
	if v.showClient {
		cv := version.New()
		fmt.Fprintf(v.out, "Client: %s\n", formatVersion(cv, v.short))
	}

	if !v.showServer {
		return nil
	}

	sv, err := v.client.Version()
	if err != nil {
		log.Debug(err)
		return errors.New("cannot connect to draftd")
	}
	fmt.Fprintf(v.out, "Server: %s\n", formatVersion(sv, v.short))
	return nil
}

func formatVersion(v *version.Version, short bool) string {
	if short {
		return fmt.Sprintf("%s+g%s", v.SemVer, v.GitCommit[:7])
	}
	return fmt.Sprintf("%#v", v)
}
