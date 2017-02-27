/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/kube"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"

	"github.com/deis/prow/cmd/prow/installer"
	"github.com/deis/prow/pkg/prow/prowpath"
	"github.com/deis/prow/pkg/prowd/portforwarder"
)

const initDesc = `
This command installs Prowd (the prow server side component) onto your
Kubernetes Cluster and sets up local configuration in $PROW_HOME (default ~/.prow/)

To set up just a local environment, use '--client-only'. That will configure
$PROW_HOME, but not attempt to connect to a remote cluster and install the Prowd
deployment.

To dump a manifest containing the Prowd deployment YAML, combine the
'--dry-run' and '--debug' flags.
`

type initCmd struct {
	image      string
	clientOnly bool
	canary     bool
	upgrade    bool
	namespace  string
	dryRun     bool
	out        io.Writer
	home       prowpath.Home
	kubeClient internalclientset.Interface
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
			i.namespace = portforwarder.ProwNamespace
			i.home = prowpath.Home(homePath())
			return i.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&i.image, "prowd-image", "i", "", "override prowd image")
	f.BoolVar(&i.canary, "canary-image", false, "use the canary image")
	f.BoolVar(&i.upgrade, "upgrade", false, "upgrade if prowd is already installed")
	f.BoolVarP(&i.clientOnly, "client-only", "c", true, "if set does not install prowd")
	f.BoolVar(&i.dryRun, "dry-run", false, "do not install local or remote")

	return cmd
}

// runInit initializes local config and installs tiller to Kubernetes Cluster
func (i *initCmd) run() error {
	if flagDebug {
		dm, err := installer.DeploymentManifest(i.namespace, i.image, i.canary)
		if err != nil {
			return err
		}
		fm := fmt.Sprintf("apiVersion: extensions/v1beta1\nkind: Deployment\n%s", dm)
		fmt.Fprintln(i.out, fm)

		sm, err := installer.ServiceManifest(i.namespace)
		if err != nil {
			return err
		}
		fm = fmt.Sprintf("apiVersion: v1\nkind: Service\n%s", sm)
		fmt.Fprintln(i.out, fm)
	}

	if i.dryRun {
		return nil
	}

	if err := ensureDirectories(i.home, i.out); err != nil {
		return err
	}
	fmt.Fprintf(i.out, "$PROW_HOME has been configured at %s.\n", prowHome)

	if !i.clientOnly {
		if i.kubeClient == nil {
			_, c, err := getKubeClient(kubeContext)
			if err != nil {
				return fmt.Errorf("could not get kubernetes client: %s", err)
			}
			i.kubeClient = c
		}
		if err := installer.Install(i.kubeClient, i.namespace, i.image, i.canary, flagDebug); err != nil {
			if !kerrors.IsAlreadyExists(err) {
				return fmt.Errorf("error installing: %s", err)
			}
			if i.upgrade {
				if err := installer.Upgrade(i.kubeClient, i.namespace, i.image, i.canary); err != nil {
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

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*restclient.Config, *internalclientset.Clientset, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return config, client, nil
}
