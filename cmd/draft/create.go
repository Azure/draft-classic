package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/technosophos/moniker"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/deis/draft/pkg/draft/draftpath"
	"github.com/deis/draft/pkg/draft/manifest"
	"github.com/deis/draft/pkg/draft/pack"
	"github.com/deis/draft/pkg/osutil"
)

const (
	createDesc = `This command transforms the local directory to be deployable via 'draft up'.
`
)

type createCmd struct {
	appName string
	out     io.Writer
	pack    string
	home    draftpath.Home
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "transform the local directory to be deployable to Kubernetes",
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run()
		},
	}

	cc.home = draftpath.Home(homePath())

	f := cmd.Flags()
	f.StringVarP(&cc.appName, "app", "a", "", "name of the Helm release. By default this is the basename of the current working directory")
	f.StringVarP(&cc.pack, "pack", "p", "", "the named Draft starter pack to scaffold the app with")

	return cmd
}

func (c *createCmd) run() error {
	var err error
	if c.appName == "" {
		c.appName = generateName()
	}

	cfile := &chart.Metadata{
		Name:        c.appName,
		Description: "A Helm chart for Kubernetes",
		Version:     "0.1.0",
		ApiVersion:  chartutil.ApiVersionV1,
	}

	chartExists, err := osutil.Exists("chart")
	if err != nil {
		return fmt.Errorf("there was an error checking if a chart exists: %v", err)
	}
	if chartExists {
		// chart dir already exists, so we just tell the user that we are happily skipping the
		// process.
		fmt.Fprintln(c.out, "--> chart/ already exists. Ready to sail!")
		return nil
	}

	if c.pack != "" {
		// --pack was explicitly defined, so we can just lazily use that here. No detection required.
		lpack := filepath.Join(c.home.Packs(), c.pack)
		err = pack.CreateFrom(cfile, "", lpack)
		if err != nil {
			return err
		}
	} else {
		// pack detection time
		packPath, output, err := doPackDetection(c.home.Packs(), c.out)
		log.Debugf("doPackDetection result: %s, %s, %v", packPath, output, err)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "--> %s app detected\n", output)
		err = pack.CreateFrom(cfile, "", packPath)
		if err != nil {
			return err
		}
	}

	// write metadata to draft.yaml
	mfest := manifest.New()
	mfest.Environments[defaultEnvironmentName] = &manifest.Environment{
		AppName: c.appName,
	}
	data, err := yaml.Marshal(mfest)
	if err != nil {
		return fmt.Errorf("could not marshal draft metadata to yaml: %v", err)
	}
	if err = ioutil.WriteFile("draft.yaml", data, 0644); err != nil {
		return fmt.Errorf("could not write metadata to draft.yaml: %v", err)
	}

	fmt.Fprintln(c.out, "--> Ready to sail")
	return nil
}

// doPackDetection performs pack detection across all the packs available in $(draft home)/packs in
// alphabetical order, returning the pack dirpath, the "formal name" returned from the detect
// script's output and any errors that occurred during the pack detection.
func doPackDetection(packHomeDir string, out io.Writer) (string, string, error) {
	files, err := ioutil.ReadDir(packHomeDir)
	if err != nil {
		return "", "", fmt.Errorf("there was an error reading %s: %v", packHomeDir, err)
	}
	for _, file := range files {
		if file.IsDir() {
			packPath := filepath.Join(packHomeDir, file.Name())
			log.Debugf("pack path: %s", packPath)
			p, err := pack.FromDir(packPath)
			if err != nil {
				return "", "", fmt.Errorf("could not load pack %s: %v", packPath, err)
			}
			output, err := p.Detect("")
			log.Debugf("pack.Detect() result: %s, %v", output, err)
			if err == nil {
				return packPath, output, err
			}
		}
	}
	return "", "", fmt.Errorf("Unable to select a starter pack Q_Q")
}

// generateName generates a random name
func generateName() string {
	namer := moniker.New()
	return namer.NameSep("-")
}
