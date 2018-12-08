package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Azure/go-autorest/autorest"
	azurecli "github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	dockerdebug "github.com/docker/cli/cli/debug"
	dockerflags "github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"k8s.io/client-go/rest"

	"github.com/Azure/draft/pkg/azure/containerregistry"
	"github.com/Azure/draft/pkg/azure/iam"
	"github.com/Azure/draft/pkg/builder"
	azurecontainerbuilder "github.com/Azure/draft/pkg/builder/azure"
	dockercontainerbuilder "github.com/Azure/draft/pkg/builder/docker"
	"github.com/Azure/draft/pkg/cmdline"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/local"
	"github.com/Azure/draft/pkg/storage/kube/configmap"
	"github.com/Azure/draft/pkg/tasks"
)

const upDesc = `
This command builds a container image using Docker, pushes it to a container registry
and then instructs helm to install the chart, referencing the image just built.
`

const (
	ignoreFileName        = ".draftignore"
	dockerTLSEnvVar       = "DOCKER_TLS"
	dockerTLSVerifyEnvVar = "DOCKER_TLS_VERIFY"
	tasksTOMLFile         = ".draft-tasks.toml"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
	autoConnect    bool
	skipImagePush  bool
	quiet          bool
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

func defaultDockerTLS() bool {
	return os.Getenv(dockerTLSEnvVar) != ""
}

func defaultDockerTLSVerify() bool {
	return os.Getenv(dockerTLSVerifyEnvVar) != ""
}

func dockerPreRun(opts *dockerflags.ClientOptions) {
	dockerflags.SetLogLevel(opts.Common.LogLevel)

	if opts.ConfigDir != "" {
		cliconfig.SetDir(opts.ConfigDir)
	}

	if opts.Common.Debug {
		dockerdebug.Enable()
	}
}

func newUpCmd(out io.Writer) *cobra.Command {
	var (
		up = &upCmd{
			out:                 out,
			dockerClientOptions: dockerflags.NewClientOptions(),
		}
		runningEnvironment string
		f                  *pflag.FlagSet
	)

	cmd := &cobra.Command{
		Use:   "up [path]",
		Short: "build and push Docker image, then install the Helm chart, referencing the image just built",
		Long:  upDesc,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			rootCmd.PersistentPreRunE(c, args)
			up.dockerClientOptions.Common.SetDefaultOptions(f)
			dockerPreRun(up.dockerClientOptions)
		},
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

	f = cmd.Flags()
	f.StringVarP(&runningEnvironment, environmentFlagName, environmentFlagShorthand, defaultDraftEnvironment(), environmentFlagUsage)
	f.BoolVar(&up.dockerClientOptions.Common.Debug, "docker-debug", false, "enable debug mode")
	f.StringVar(&up.dockerClientOptions.Common.LogLevel, "docker-log-level", "info", `set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	f.BoolVar(&up.dockerClientOptions.Common.TLS, "docker-tls", defaultDockerTLS(), "use TLS; implied by --tlsverify")
	f.BoolVar(&up.dockerClientOptions.Common.TLSVerify, fmt.Sprintf("docker-%s", dockerflags.FlagTLSVerify), defaultDockerTLSVerify(), "use TLS and verify the remote")
	f.StringVar(&up.dockerClientOptions.ConfigDir, "docker-config", cliconfig.Dir(), "location of client config files")
	f.BoolVarP(&autoConnect, "auto-connect", "", false, "specifies if draft up should automatically connect to the application")
	f.BoolVar(&skipImagePush, "skip-image-push", false, "skip pushing image to registry")
	f.BoolVarP(&quiet, "quiet", "q", false, "only output errors")

	up.dockerClientOptions.Common.TLSOptions = &tlsconfig.Options{
		CAFile:   filepath.Join(dockerCertPath, dockerflags.DefaultCaFile),
		CertFile: filepath.Join(dockerCertPath, dockerflags.DefaultCertFile),
		KeyFile:  filepath.Join(dockerCertPath, dockerflags.DefaultKeyFile),
	}

	tlsOptions := up.dockerClientOptions.Common.TLSOptions
	f.Var(opts.NewQuotedString(&tlsOptions.CAFile), "docker-tlscacert", "Trust certs signed only by this CA")
	f.Var(opts.NewQuotedString(&tlsOptions.CertFile), "docker-tlscert", "Path to TLS certificate file")
	f.Var(opts.NewQuotedString(&tlsOptions.KeyFile), "docker-tlskey", "Path to TLS key file")

	hostOpt := opts.NewNamedListOptsRef("docker-hosts", &up.dockerClientOptions.Common.Hosts, opts.ValidateHost)
	f.Var(hostOpt, "docker-host", "Daemon socket(s) to connect to")

	return cmd
}

func (u *upCmd) run(environment string) (err error) {
	var (
		buildctx   *builder.Context
		kubeConfig *rest.Config
		ctx        = context.Background()
		bldr       = builder.New()
	)
	bldr.LogsDir = u.home.Logs()

	taskList, err := tasks.Load(tasksTOMLFile)
	if err != nil {
		if err == tasks.ErrNoTaskFile {
			debug(err.Error())
		} else {
			return err
		}
	} else {
		if _, err = taskList.Run(tasks.DefaultRunner, tasks.PreUp, ""); err != nil {
			return err
		}
	}

	if buildctx, err = builder.LoadWithEnv(u.src, environment); err != nil {
		return fmt.Errorf("failed loading build context with env %q: %v", environment, err)
	}

	if configuredBuilder, ok := globalConfig[containerBuilder.name]; ok {
		buildctx.Env.ContainerBuilder = configuredBuilder
	}

	// if a registry has been set in their global config but nothing was in draft.toml, use that instead.
	if reg, ok := globalConfig[registry.name]; ok {
		buildctx.Env.Registry = reg
	}

	// Check if skip-image-push is specified. If so, unset registry.
	if skipImagePush {
		buildctx.Env.Registry = ""
	}

	if configuredResourceGroup, ok := globalConfig[resourceGroupName.name]; ok {
		buildctx.Env.ResourceGroupName = configuredResourceGroup
	}

	if buildctx.Env.Registry == "" && !skipImagePush {
		// give a way for minikube users (and users who understand what they're doing) a way to opt out
		if _, ok := globalConfig[disablePushWarning.name]; !ok {
			fmt.Fprintln(u.out, "WARNING: no registry has been set, therefore Draft will not push to a container registry. This can be fixed by running `draft config set registry docker.io/myusername`")
			fmt.Fprintln(u.out, "Hint: this warning can be disabled by running `draft config set disable-push-warning 1`")
		}
	}

	var cb builder.ContainerBuilder
	switch buildctx.Env.ContainerBuilder {
	case "acrbuild":
		subscription, err := getSubscriptionFromProfile()
		if err != nil {
			return fmt.Errorf("Could not retrieve azure profile information: %v", err)
		}
		token, err := iam.GetToken(iam.AuthGrantType())
		if err != nil {
			return fmt.Errorf("Could not retrieve adal token: %v", err)
		}
		auth := autorest.NewBearerAuthorizer(&token)
		registriesClient := containerregistry.NewRegistriesClient(subscription.ID)
		registriesClient.Authorizer = auth
		registriesClient.AddToUserAgent(containerregistry.UserAgent())
		buildsClient := containerregistry.NewBuildsClient(subscription.ID)
		buildsClient.Authorizer = auth
		buildsClient.AddToUserAgent(containerregistry.UserAgent())
		cb = &azurecontainerbuilder.Builder{
			RegistryClient: registriesClient,
			BuildsClient:   buildsClient,
			AdalToken:      token,
			Subscription:   subscription,
		}
	default:
		// setup docker
		cli := &command.DockerCli{}
		if err := cli.Initialize(u.dockerClientOptions); err != nil {
			return fmt.Errorf("failed to create docker client: %v", err)
		}
		cb = &dockercontainerbuilder.Builder{
			DockerClient: cli,
		}
	}
	bldr.ContainerBuilder = cb

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
	progressC := bldr.Up(ctx, buildctx)
	opts := []cmdline.Option{cmdline.WithBuildID(bldr.ID)}

	if quiet {
		opts = append(opts, cmdline.WithStdout(ioutil.Discard))
	}

	if displayEmoji {
		opts = append(opts, cmdline.WithDisplayEmoji(displayEmoji))
	}

	cmdline.Display(ctx, buildctx.Env.Name, progressC, opts...)

	if buildctx.Env.AutoConnect || autoConnect {
		c := newConnectCmd(u.out)
		return c.RunE(c, []string{})
	}

	if err := runPostDeployTasks(taskList, bldr.ID); err != nil {
		debug(err.Error())
	}

	if _, err = taskList.Run(tasks.DefaultRunner, tasks.PostUp, ""); err != nil {
		debug(err.Error())
	}

	return nil
}

func runPostDeployTasks(taskList *tasks.Tasks, buildID string) error {
	if taskList == nil || len(taskList.PostDeploy) == 0 {
		return errors.New("No post deploy tasks to run")
	}

	app, err := local.DeployedApplication(draftToml, runningEnvironment)
	if err != nil {
		return err
	}

	client, _, err := getKubeClient(kubeContext)
	if err != nil {
		return err
	}

	names, err := app.GetPodNames(buildID, client)
	if err != nil {
		return err
	}

	for _, name := range names {
		_, err := taskList.Run(tasks.DefaultRunner, tasks.PostDeploy, name)
		if err != nil {
			debug("error running task: %v", err)
		}
	}

	return nil
}

func getSubscriptionFromProfile() (azurecli.Subscription, error) {
	profilePath, err := azurecli.ProfilePath()
	if err != nil {
		return azurecli.Subscription{}, err
	}
	profile, err := azurecli.LoadProfile(profilePath)
	if err != nil {
		return azurecli.Subscription{}, err
	}
	for _, sub := range profile.Subscriptions {
		if sub.IsDefault {
			return sub, nil
		}
	}
	return azurecli.Subscription{}, fmt.Errorf("could not find a default subscription ID from %s", profilePath)
}
