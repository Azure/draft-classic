package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Azure/draft/pkg/builder"
	"github.com/Azure/draft/pkg/cmdline"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/storage/kube/configmap"
	"github.com/docker/cli/cli/command"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"k8s.io/client-go/rest"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to
the draft server.

Adding the "watch" option to draft.toml makes draft automatically archive and
upload whenever local files are saved. Draft delays a couple seconds to ensure
that changes have stopped before uploading, but that can be altered by the
"watch-delay" option.
`

const (
	ignoreFileName = ".draftignore"
)

type upCmd struct {
	out  io.Writer
	src  string
	home draftpath.Home
	// storage engine draft should use for storing builds, logs, etc.
	storageEngine string
}

func newUpCmd(out io.Writer) *cobra.Command {
	var (
		up                 = &upCmd{out: out}
		runningEnvironment string
	)

	cmd := &cobra.Command{
		Use:   "up [path]",
		Short: "upload the current directory to the draft server for deployment",
		Long:  upDesc,
		RunE: func(_ *cobra.Command, args []string) (err error) {
			if len(args) > 0 {
				up.src = args[0]
			}
			if up.src == "" || up.src == "." {
				if up.src, err = os.Getwd(); err != nil {
					return err
				}
			}
			up.home = draftpath.Home(homePath())
			return up.run(runningEnvironment)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&runningEnvironment, environmentFlagName, environmentFlagShorthand, defaultDraftEnvironment(), environmentFlagUsage)

	return cmd
}

func (u *upCmd) run(environment string) (err error) {
	var (
		buildctx   *builder.Context
		kubeConfig *rest.Config
		ctx        = context.Background()
		bldr       = &builder.Builder{
			LogsDir: u.home.Logs(),
		}
	)
	if buildctx, err = builder.LoadWithEnv(u.src, environment); err != nil {
		return fmt.Errorf("failed loading build context with env %q: %v", environment, err)
	}

	// if a registry has been set in their global config but nothing was in draft.toml, use that instead
	if reg, ok := globalConfig["registry"]; ok {
		buildctx.Env.Registry = reg
	}

	// setup docker
	cli := &command.DockerCli{}
	if err := cli.Initialize(dockerflags.NewClientOptions()); err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}
	bldr.DockerClient = cli

	// setup kube
	bldr.Kube, kubeConfig, err = getKubeClient(kubeContext)
	if err != nil {
		return fmt.Errorf("Could not get a kube client: %s", err)
	}
	bldr.Helm, err = setupHelm(bldr.Kube, kubeConfig, tillerNamespace)
	if err != nil {
		return fmt.Errorf("Could not get a helm client: %s", err)
	}

	// setup the storage engine
	bldr.Storage = configmap.NewConfigMaps(bldr.Kube.CoreV1().ConfigMaps(tillerNamespace))

	cmdline.Display(ctx, buildctx.Env.Name, bldr.Up(ctx, buildctx))
	return nil
}
