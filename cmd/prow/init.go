package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/strvals"
	"k8s.io/helm/pkg/tiller/environment"
	kerrors "k8s.io/kubernetes/pkg/api/errors"

	"github.com/deis/prow/cmd/prow/installer"
	"github.com/deis/prow/pkg/prow/pack"
	"github.com/deis/prow/pkg/prow/prowpath"
)

const initDesc = `
This command installs Prowd (the prow server side component) onto your
Kubernetes Cluster and sets up local configuration in $PROW_HOME (default ~/.prow/)

To set up just a local environment, use '--client-only'. That will configure
$PROW_HOME, but not attempt to connect to a remote cluster and install the Prowd
deployment.

To dump information about the Prowd chart, combine the '--dry-run' and '--debug' flags.
`

type initCmd struct {
	clientOnly        bool
	upgrade           bool
	dryRun            bool
	out               io.Writer
	home              prowpath.Home
	helmClient        *helm.Client
	values            []string
	rawValueFilePaths []string
	tillerNamespace   string
}

func newInitCmd(out io.Writer) *cobra.Command {
	i := &initCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize Prow on both client and server",
		Long:  initDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("This command does not accept arguments")
			}
			i.home = prowpath.Home(homePath())
			return i.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&i.tillerNamespace, "tiller-namespace", environment.DefaultTillerNamespace, "the namespace tiller is deployed to. This will also be where prowd is deployed to.")
	f.StringArrayVar(&i.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVarP(&i.rawValueFilePaths, "values", "f", []string{}, "specify prowd values in a YAML file (can specify multiple)")
	f.BoolVar(&i.upgrade, "upgrade", false, "upgrade if prowd is already installed")
	f.BoolVarP(&i.clientOnly, "client-only", "c", false, "if set does not install prowd")
	f.BoolVar(&i.dryRun, "dry-run", false, "do not install local or remote")

	return cmd
}

func (i *initCmd) vals() ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range i.rawValueFilePaths {
		currentMap := map[string]interface{}{}
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range i.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// runInit initializes local config and installs Prowd to Kubernetes Cluster
func (i *initCmd) run() error {
	chartConfig := new(chart.Config)

	rawVals, err := i.vals()
	if err != nil {
		return err
	}
	chartConfig.Raw = string(rawVals)

	if flagDebug {
		chart, err := chartutil.LoadFiles(installer.DefaultChartFiles)
		if err != nil {
			return err
		}
		fmt.Fprintln(i.out, chart)
	}

	if i.dryRun {
		return nil
	}

	if err := ensureDirectories(i.home, i.out); err != nil {
		return err
	}
	if err := ensurePacks(i.home, i.out); err != nil {
		return err
	}
	fmt.Fprintf(i.out, "$PROW_HOME has been configured at %s.\n", prowHome)

	if !i.clientOnly {
		if i.helmClient == nil {
			clientset, config, err := getKubeClient(kubeContext)
			if err != nil {
				return fmt.Errorf("Could not get a kube client: %s", err)
			}
			tunnel, err := portforwarder.New(i.tillerNamespace, clientset, config)
			if err != nil {
				return fmt.Errorf("Could not get a connection to tiller: %s", err)
			}
			i.helmClient = helm.NewClient(helm.Host(fmt.Sprintf("localhost:%d", tunnel.Local)))
		}

		if err := installer.Install(i.helmClient, chartConfig, i.tillerNamespace); err != nil {
			if !kerrors.IsAlreadyExists(err) {
				return fmt.Errorf("error installing: %s", err)
			}
			if i.upgrade {
				if err := installer.Upgrade(i.helmClient, chartConfig); err != nil {
					return fmt.Errorf("error when upgrading: %s", err)
				}
				fmt.Fprintln(i.out, "\nProwd (the prow server side component) has been upgraded to the current version.")
			} else {
				fmt.Fprintln(i.out, "Warning: Prowd is already installed in the cluster.\n"+
					"(Use --client-only to suppress this message, or --upgrade to upgrade Prowd to the current version.)")
			}
		} else {
			fmt.Fprintln(i.out, "\nProwd (the prow server side component) has been installed into your Kubernetes Cluster.")
		}
	} else {
		fmt.Fprintln(i.out, "Not installing Prowd due to 'client-only' flag having been set")
	}

	fmt.Fprintln(i.out, "Happy Sailing!")
	return nil
}

// ensureDirectories checks to see if $PROW_HOME exists
//
// If $PROW_HOME does not exist, this function will create it.
func ensureDirectories(home prowpath.Home, out io.Writer) error {
	configDirectories := []string{
		home.String(),
		home.Plugins(),
		home.Packs(),
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			fmt.Fprintf(out, "Creating %s \n", p)
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
func ensurePacks(home prowpath.Home, out io.Writer) error {
	packNames := []string{"zzznginx"}
	for _, packName := range packNames {
		fmt.Fprintf(out, "Creating pack %s\n", packName)
		if _, err := pack.Create(packName, home.Packs()); err != nil {
			return err
		}
	}
	return nil
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
