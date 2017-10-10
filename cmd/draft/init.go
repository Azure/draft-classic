package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"

	"github.com/Azure/draft/cmd/draft/installer"
	installerConfig "github.com/Azure/draft/cmd/draft/installer/config"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/pack/repo"
	"github.com/Azure/draft/pkg/plugin"
)

const (
	initDesc = `
This command installs the server side component of Draft onto your
Kubernetes Cluster and sets up local configuration in $DRAFT_HOME (default ~/.draft/)

To set up just a local environment, use '--client-only'. That will configure
$DRAFT_HOME, but not attempt to connect to a remote cluster and install the Draft
deployment.

To dump information about the Draft chart, combine the '--dry-run' and '--debug' flags.
`
	chartConfigTpl = `
ingress:
  enabled: %s
  basedomain: %s
registry:
  url: %s
  authtoken: %s
imageOverride: %s
`
)

type initCmd struct {
	clientOnly     bool
	out            io.Writer
	in             io.Reader
	home           draftpath.Home
	autoAccept     bool
	helmClient     *helm.Client
	ingressEnabled bool
	image          string
}

func newInitCmd(out io.Writer, in io.Reader) *cobra.Command {
	i := &initCmd{
		out: out,
		in:  in,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize Draft on both client and server",
		Long:  initDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("This command does not accept arguments")
			}
			i.home = draftpath.Home(homePath())
			return i.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&i.clientOnly, "client-only", "c", false, "install local configuration, but skip remote configuration")
	f.BoolVarP(&i.ingressEnabled, "ingress-enabled", "", false, "configure ingress")
	f.BoolVar(&i.autoAccept, "auto-accept", false, "automatically accept configuration defaults (if detected). It will still prompt for information if this is set to true and no cloud provider was found")
	f.StringVarP(&i.image, "draftd-image", "i", "", "override Draftd image")

	return cmd
}

// runInit initializes local config and installs Draft to Kubernetes Cluster
func (i *initCmd) run() error {

	if err := i.setupDraftHome(); err != nil {
		return err
	}
	fmt.Fprintf(i.out, "$DRAFT_HOME has been configured at %s.\n", draftHome)

	if !i.clientOnly {
		client, clientConfig, err := getKubeClient(kubeContext)
		if err != nil {
			return fmt.Errorf("Could not get a kube client: %s", err)
		}
		restClientConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return fmt.Errorf("Could not retrieve client config from the kube client: %s", err)
		}

		i.helmClient, err = setupHelm(client, restClientConfig, draftNamespace)
		if err != nil {
			return err
		}

		draftConfig, cloudProvider, err := installerConfig.FromClientConfig(clientConfig)
		if err != nil {
			return fmt.Errorf("Could not generate chart config from kube client config: %s", err)
		}

		if cloudProvider == "minikube" {
			fmt.Fprintf(i.out, "\nDraft detected that you are using %s as your cloud provider. AWESOME!\n", cloudProvider)

			if !i.autoAccept {
				fmt.Fprint(i.out, "Is it okay to use the registry addon in minikube to store your application images?\nIf not, we will prompt you for information on the registry you'd like to push your application images to during development. [Y/n] ")
				reader := bufio.NewReader(i.in)
				text, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("Could not read input: %s", err)
				}
				text = strings.TrimSpace(text)
				if text == "" || strings.ToLower(text) == "y" {
					i.autoAccept = true
				}
			}
		}

		draftConfig.Ingress = i.ingressEnabled
		draftConfig.Image = i.image

		if !i.autoAccept || cloudProvider == "" {
			// prompt for missing information
			fmt.Fprintf(i.out, "\nIn order to configure Draft, we need a bit more information...\n\n")

			reader := bufio.NewReader(i.in)
			if err := setupContainerRegistry(i.out, reader, draftConfig); err != nil {
				return err
			}

			if err := setupBasedomain(i.out, reader, i.ingressEnabled, draftConfig); err != nil {
				return err
			}

		}

		rawChartConfig := fmt.Sprintf(chartConfigTpl, strconv.FormatBool(draftConfig.Ingress), draftConfig.Basedomain, draftConfig.RegistryURL, draftConfig.RegistryAuth, draftConfig.Image)
		// attempt to purge the old release, but log errors to --debug
		if err := installer.Uninstall(i.helmClient); err != nil {
			log.Debugf("error uninstalling Draft: %s", err)
		}
		if err := installer.Install(i.helmClient, draftNamespace, rawChartConfig); err != nil {
			return fmt.Errorf("error installing Draft: %s", err)
		}
		fmt.Fprintln(i.out, "Draft has been installed into your Kubernetes Cluster.")
	} else {
		fmt.Fprintln(i.out, "Skipped installing Draft's server side component in Kubernetes due to 'client-only' flag having been set")
	}

	fmt.Fprintln(i.out, "Happy Sailing!")
	return nil
}

// ensureDirectories checks to see if $DRAFT_HOME exists
//
// If $DRAFT_HOME does not exist, this function will create it.
func (i *initCmd) ensureDirectories() error {
	configDirectories := []string{
		i.home.String(),
		i.home.Plugins(),
		i.home.Packs(),
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			fmt.Fprintf(i.out, "Creating %s \n", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}

// ensurePacks checks to see if the default packs exist.
//
// If the pack does not exist, this function will create it.
func (i *initCmd) ensurePacks() error {
	for _, builtin := range repo.Builtins() {
		if err := i.ensurePack(builtin); err != nil {
			return err
		}
	}
	return nil
}

func (i *initCmd) ensurePack(builtin *repo.Builtin) error {
	addArgs := []string{
		"add",
		builtin.URL,
		"--version",
		builtin.Version,
	}
	packRepoCmd, _, err := newRootCmd(i.out, i.in).Find([]string{"pack-repo"})
	if err != nil {
		return err
	}
	if err := packRepoCmd.RunE(packRepoCmd, addArgs); err != nil {
		fmt.Fprintf(i.out, "Removing pack %s then re-trying...\n", builtin.Name)
		// remove repo, then re-install
		var builtinRepo *repo.Repository
		for _, repo := range repo.FindRepositories(i.home.Packs()) {
			if repo.Name == builtin.Name {
				builtinRepo = &repo
			}
		}
		if builtinRepo != nil {
			if removeErr := os.RemoveAll(builtinRepo.Dir); removeErr != nil {
				return removeErr
			}
		}
		return packRepoCmd.RunE(packRepoCmd, addArgs)
	}
	return nil
}

// IsReleaseAlreadyExists returns true if err matches the "release already exists"
// error from Helm; else returns false
func IsReleaseAlreadyExists(err error) bool {
	alreadyExistsRegExp := regexp.MustCompile("a release named \"(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])+\" already exists")
	return alreadyExistsRegExp.MatchString(err.Error())
}

// ensurePlugins checks to see if the default plugins exist.
//
// If the plugin does not exist, this function will add it.
func (i *initCmd) ensurePlugins() error {
	for _, builtin := range plugin.Builtins() {
		if err := ensurePlugin(i.out, i.in, builtin); err != nil {
			return err
		}
	}
	return nil
}

func ensurePlugin(out io.Writer, in io.Reader, builtin *plugin.Builtin) error {
	var (
		installArgs = []string{
			os.Args[0],
			"plugin",
			"install",
			builtin.URL,
			"--version",
			builtin.Version,
		}
		removeArgs = []string{
			os.Args[0],
			"plugin",
			"remove",
			builtin.Name,
		}
	)

	fmt.Fprintf(out, "Adding plugin %s...\n", builtin.URL)
	os.Args = installArgs
	cmd := newRootCmd(out, in)
	err := cmd.Execute()
	if err == plugin.ErrExists {
		// remove plugin, then re-install
		os.Args = removeArgs
		if removeErr := cmd.Execute(); removeErr != nil {
			return removeErr
		}
		os.Args = installArgs
		return cmd.Execute()
	}
	return err
}

func (i *initCmd) setupDraftHome() error {
	ensureFuncs := []func() error{
		i.ensureDirectories,
		i.ensurePlugins,
		i.ensurePacks,
	}

	for _, funct := range ensureFuncs {
		if err := funct(); err != nil {
			return err
		}
	}

	return nil
}

func setupTillerConnection(client kubernetes.Interface, restClientConfig *restclient.Config, namespace string) (*kube.Tunnel, error) {
	tunnel, err := portforwarder.New(namespace, client, restClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Could not get a connection to tiller: %s\nPlease ensure you have run `helm init`", err)
	}

	return tunnel, err
}

func setupBasedomain(out io.Writer, reader *bufio.Reader, ingress bool, draftConfig *installerConfig.DraftConfig) error {
	if ingress {
		fmt.Fprint(out, "4. Enter your top-level domain for ingress (e.g. draft.example.com): ")
		basedomain, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("Could not read input: %s", err)
		}
		draftConfig.Basedomain = strings.TrimSpace(basedomain)
	} else {
		draftConfig.Basedomain = ""
	}

	return nil
}

func setupHelm(kubeClient *kubernetes.Clientset, restClientConfig *restclient.Config, namespace string) (*helm.Client, error) {
	tunnel, err := setupTillerConnection(kubeClient, restClientConfig, namespace)
	if err != nil {
		return nil, err
	}

	return helm.NewClient(helm.Host(fmt.Sprintf("localhost:%d", tunnel.Local))), nil
}

func setupContainerRegistry(out io.Writer, reader *bufio.Reader, draftConfig *installerConfig.DraftConfig) error {
	fmt.Fprint(out, "1. Enter your Docker registry URL (e.g. docker.io/myuser, quay.io/myuser, myregistry.azurecr.io): ")
	registryURL, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Could not read input: %s", err)
	}
	draftConfig.RegistryURL = strings.TrimSpace(registryURL)

	fmt.Fprint(out, "2. Enter your username: ")
	dockerUser, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Could not read input: %s", err)
	}
	dockerUser = strings.TrimSpace(dockerUser)
	fmt.Fprint(out, "3. Enter your password: ")
	// NOTE(bacongobbler): casting syscall.Stdin here to an int is intentional here as on
	// Windows, syscall.Stdin is a Handler, which is of type uintptr.
	dockerPass, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return fmt.Errorf("Could not read input: %s", err)
	}

	registryAuth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf(
			`{"username":"%s","password":"%s"}`,
			dockerUser,
			dockerPass)))

	draftConfig.RegistryAuth = registryAuth
	return nil
}
