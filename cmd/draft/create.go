package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/manifest"
	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/Azure/draft/pkg/osutil"
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
	dest    string
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "create [path]",
		Short: "transform the local directory to be deployable to Kubernetes",
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cc.dest = args[0]
			}
			return cc.run()
		},
	}

	cc.home = draftpath.Home(homePath())

	f := cmd.Flags()
	f.StringVarP(&cc.appName, "app", "a", "", "name of the Helm release. By default, this is a randomly generated name")
	f.StringVarP(&cc.pack, "pack", "p", "", "the named Draft starter pack to scaffold the app with")

	return cmd
}

func (c *createCmd) run() error {
	var err error
	mfest := manifest.New()

	if c.appName != "" {
		mfest.Environments[manifest.DefaultEnvironmentName].Name = c.appName
	}

	cfile := &chart.Metadata{
		Name:        mfest.Environments[manifest.DefaultEnvironmentName].Name,
		Description: "A Helm chart for Kubernetes",
		Version:     "0.1.0",
		ApiVersion:  chartutil.ApiVersionV1,
	}

	chartExists, err := osutil.Exists(filepath.Join(c.dest, "chart"))
	if err != nil {
		return fmt.Errorf("there was an error checking if a chart exists: %v", err)
	}
	if chartExists {
		// chart dir already exists, so we just tell the user that we are happily skipping the
		// process.
		fmt.Fprintln(c.out, "--> chart directory already exists. Ready to sail!")
		return nil
	}

	if c.pack != "" {
		// --pack was explicitly defined, so we can just lazily use that here. No detection required.
		lpack := filepath.Join(c.home.Packs(), c.pack)
		err = pack.CreateFrom(cfile, c.dest, lpack)
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
		err = pack.CreateFrom(cfile, c.dest, packPath)
		if err != nil {
			return err
		}
	}
	tomlFile := filepath.Join(c.dest, "draft.toml")
	draftToml, err := os.OpenFile(tomlFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer draftToml.Close()

	if err := toml.NewEncoder(draftToml).Encode(mfest); err != nil {
		return fmt.Errorf("could not write metadata to draft.toml: %v", err)
	}

	ignoreFile := filepath.Join(c.dest, ignoreFileName)
	if _, err := os.Stat(ignoreFile); os.IsNotExist(err) {
		d1 := []byte("*.swp\n*.tmp\n*.temp\n.git*\n")
		if err := ioutil.WriteFile(ignoreFile, d1, 0644); err != nil {
			return err
		}
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
