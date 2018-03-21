package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Azure/draft/pkg/builder"
	"github.com/Azure/draft/pkg/cmdline"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/storage/kube/configmap"
	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
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
	ignoreFileName        = ".draftignore"
	dockerTLSEnvVar       = "DOCKER_TLS"
	dockerTLSVerifyEnvVar = "DOCKER_TLS_VERIFY"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

type upCmd struct {
	out  io.Writer
	src  string
	home draftpath.Home
	// storage engine draft should use for storing builds, logs, etc.
	storageEngine string
	// options common to the docker client and the daemon.
	dockerClientOptions *dockerflags.ClientOptions
}

func stringPtr(s string) *string { return &s }

func defaultDockerTLS() bool {
	return os.Getenv(dockerTLSEnvVar) != ""
}

func defaultDockerTLSVerify() bool {
	return os.Getenv(dockerTLSVerifyEnvVar) != ""
}

func newUpCmd(out io.Writer) *cobra.Command {
	var (
		up = &upCmd{
			out:                 out,
			dockerClientOptions: dockerflags.NewClientOptions(),
		}
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
	f.BoolVar(&up.dockerClientOptions.Common.Debug, "docker-debug", false, "Enable debug mode")
	f.StringVar(&up.dockerClientOptions.Common.LogLevel, "docker-log-level", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	f.BoolVar(&up.dockerClientOptions.Common.TLS, "docker-tls", defaultDockerTLS(), "Use TLS; implied by --tlsverify")
	f.BoolVar(&up.dockerClientOptions.Common.TLSVerify, fmt.Sprintf("docker-%s", dockerflags.FlagTLSVerify), defaultDockerTLSVerify(), "Use TLS and verify the remote")
	f.StringVar(&up.dockerClientOptions.ConfigDir, "docker-config", cliconfig.Dir(), "Location of client config files")
	f.Var(opts.NewQuotedString(stringPtr(filepath.Join(dockerCertPath, dockerflags.DefaultCaFile))), "docker-tlscacert", "Trust certs signed only by this CA")
	f.Var(opts.NewQuotedString(stringPtr(filepath.Join(dockerCertPath, dockerflags.DefaultCertFile))), "docker-tlscert", "Path to TLS certificate file")
	f.Var(opts.NewQuotedString(stringPtr(filepath.Join(dockerCertPath, dockerflags.DefaultKeyFile))), "docker-tlskey", "Path to TLS key file")

	hostOpt := opts.NewNamedListOptsRef("docker-hosts", &up.dockerClientOptions.Common.Hosts, opts.ValidateHost)
	f.Var(hostOpt, "docker-host", "Daemon socket(s) to connect to")

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
	if err := cli.Initialize(u.dockerClientOptions); err != nil {
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
