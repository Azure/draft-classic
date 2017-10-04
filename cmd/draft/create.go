package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/manifest"
	"github.com/Azure/draft/pkg/draft/pack"
	"github.com/Azure/draft/pkg/linguist"
	"github.com/Azure/draft/pkg/osutil"
)

const (
	draftToml  = "draft.toml"
	createDesc = `This command transforms the local directory to be deployable via 'draft up'.
`
)

type createCmd struct {
	appName string
	out     io.Writer
	in      io.Reader
	pack    string
	home    draftpath.Home
	dest    string
}

func newCreateCmd(out io.Writer, in io.Reader) *cobra.Command {
	cc := &createCmd{
		out: out,
		in:  in,
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
		err = pack.CreateFrom(c.dest, lpack)
		if err != nil {
			return err
		}
	} else {
		// pack detection time
		var selectedPack *pack.Pack
		detectedPacks, err := doPackDetection(c.home.Packs(), c.out)
		if err != nil {
			return err
		}

		for i := range detectedPacks {
			log.Debugln("detected pack %s", detectedPacks[i].Name)
		}
		switch len(detectedPacks) {
		case 0:
			return errors.New("Could not find a starter pack Q_Q")
		case 1:
			selectedPack = detectedPacks[0]
		default:
			for selectedPack == nil {
				fmt.Fprintf(c.out, "!!! draft detected multiple packs that satisfy the detected language. please choose from the following:\n\n")
				for i := range detectedPacks {
					fmt.Fprintf(c.out, "%d. %s\n", i, detectedPacks[i].Name)
				}
				fmt.Fprintf(c.out, "\nEnter your choice: ")
				reader := bufio.NewReader(c.in)
				text, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("Could not read input: %s", err)
				}
				text = strings.TrimSpace(text)
				selectedNumber, err := strconv.Atoi(text)
				if err != nil {
					fmt.Fprintf(c.out, "!!! Invalid choice: %s\n", text)
					continue
				}
				if selectedNumber < 0 || selectedNumber >= len(detectedPacks) {
					fmt.Fprintf(c.out, "!!! Invalid choice: %s\n", text)
					continue
				}
				selectedPack = detectedPacks[selectedNumber]
			}
		}

		if err := pack.CreateFrom(c.dest, selectedPack.Path); err != nil {
			return err
		}
	}
	tomlFile := filepath.Join(c.dest, draftToml)
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
// alphabetical order, returning the pack name and any errors that occurred during the pack detection.
func doPackDetection(packHomeDir string, out io.Writer) ([]*pack.Pack, error) {
	var packs []*pack.Pack
	langs, err := linguist.ProcessDir(".")
	log.Debugf("linguist.ProcessDir('.') result:\n\nError: %v", err)
	if err != nil {
		return nil, fmt.Errorf("there was an error detecting the language: %s", err)
	}
	for _, lang := range langs {
		log.Debugf("%s:\t%f (%s)", lang.Language, lang.Percent, lang.Color)
	}
	if len(langs) == 0 {
		return nil, errors.New("No languages were detected. Are you sure there's code in here?")
	}
	detectedLang := linguist.Alias(langs[0])
	fmt.Fprintf(out, "--> Draft detected the primary language as %s with %f%% certainty.\n", detectedLang.Language, detectedLang.Percent)
	files, err := ioutil.ReadDir(packHomeDir)
	if err != nil {
		return nil, fmt.Errorf("there was an error reading %s: %v", packHomeDir, err)
	}
	for _, packDir := range files {
		if packDir.IsDir() {
			p, err := pack.FromDir(path.Join(packHomeDir, packDir.Name()))
			if err != nil {
				return nil, fmt.Errorf("there was an error loading pack from dir %s: %v", path.Join(packHomeDir, packDir.Name()), err)
			}
			if strings.Compare(strings.ToLower(detectedLang.Language), strings.ToLower(p.Language)) == 0 {
				log.Debugf("pack path: %s", p.Path)
				packs = append(packs, p)
			}
		}
	}
	return packs, nil
}
