package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/rjeczalik/notify"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/ignore"

	"github.com/Azure/draft/pkg/draft"
	"github.com/Azure/draft/pkg/draft/manifest"
	"github.com/Azure/draft/pkg/draft/pack"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to
the draft server.

Adding the "watch" option to draft.toml makes draft automatically archive and
upload whenever local files are saved. Draft delays a couple seconds to ensure
that changes have stopped before uploading, but that can be altered by the
"watch_delay" option.
`

const (
	environmentEnvVar = "DRAFT_ENV"
	ignoreFileName    = ".draftignore"
)

type upCmd struct {
	Client   *draft.Client
	Out      io.Writer
	Manifest *manifest.Manifest
	src      string
}

func newUpCmd(out io.Writer) *cobra.Command {
	var (
		up = &upCmd{
			Out:      out,
			Manifest: manifest.New(),
		}
		runningEnvironment string
	)

	cmd := &cobra.Command{
		Use:     "up [path]",
		Short:   "upload the current directory to the draft server for deployment",
		Long:    upDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				up.src = args[0]
			}
			up.Client = ensureDraftClient(up.Client)
			up.Manifest = manifest.New()

			if up.src == "" || up.src == "." {
				var err error
				up.src, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			draftToml, err := ioutil.ReadFile(filepath.Join(up.src, "draft.toml"))
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				if err = toml.Unmarshal(draftToml, up.Manifest); err != nil {
					return fmt.Errorf("could not unmarshal draft.toml: %v", err)
				}
			}
			return up.run(runningEnvironment)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&runningEnvironment, "environment", "e", defaultDraftEnvironment(), "the environment (development, staging, qa, etc) that draft will run under")

	return cmd
}

func vals(e *manifest.Environment, cwd string) ([]byte, error) {
	base := map[string]interface{}{}

	// load $PWD/chart/values.yaml as the base config
	valuesPath := filepath.Join(cwd, pack.ChartDir, pack.ValuesfileName)

	if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
		return []byte{}, errors.New("Could not detect a proper helm chart\nTry running a `draft create` first")
	}
	bytes, err := ioutil.ReadFile(valuesPath)
	if err != nil {
		return []byte{}, err
	}
	if err := yaml.Unmarshal(bytes, &base); err != nil {
		return []byte{}, fmt.Errorf("failed to parse %s: %s", valuesPath, err)
	}

	return yaml.Marshal(base)
}

func (u *upCmd) run(environment string) (err error) {
	env := u.Manifest.Environments[environment]
	cwd := u.src
	u.Client.OptionWait = env.Wait

	rawVals, err := vals(env, cwd)
	if err != nil {
		return err
	}

	if err = u.doUp(environment, cwd, rawVals); err != nil {
		return err
	}

	// if `--watch=false`, return now
	if !env.Watch {
		return nil
	}
	fmt.Fprintln(u.Out, "Watching local files for changes...")

	notifyPath := filepath.Join(cwd, "...")
	log.Debugf("NOTIFY PATH: %s", notifyPath)
	// make a buffered channel of filesystem notification events
	ch := make(chan notify.EventInfo, 1)

	// watch the current directory and everything under it, sending events to the channel
	if err := notify.Watch(notifyPath, ch, notify.All); err != nil {
		log.Fatalf("could not watch local filesystem for changes: %v", err)
	}
	defer notify.Stop(ch)

	r, err := ignore.ParseFile(ignoreFileName)
	if err != nil {
		// only fail if file can't be parsed but exists
		if _, err := os.Stat(ignoreFileName); os.IsExist(err) {
			log.Fatalf("could not load ignore watch list %v", err)
		}
	}

	// create a timer to enforce a "quiet period" before deploying the app
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	delay := time.Duration(env.WatchDelay) * time.Second

	for {
		select {
		case evt := <-ch:
			log.Debugf("Event %s", evt)
			p := strings.TrimPrefix(evt.Path(), cwd+"/")
			fi, err := os.Stat(evt.Path())
			if err != nil {
				// create dummy file info for removed file or directory
				fi = removedFileInfo(filepath.Base(evt.Path()))
			}
			// only rebuild if the changed file isn't in our ignore list
			if r != nil && r.Ignore(p, fi) {
				continue
			}
			// ignore manually everything inside the .git/ directory as
			// helm ignore file doesn't have directory and whole content
			// (subdir of subdir) ignore support yet.
			if filepath.HasPrefix(p, ".git/") {
				continue
			}
			// reset the timer when files have changed
			timer.Reset(delay)
		case <-timer.C:
			if err = u.doUp(environment, cwd, rawVals); err != nil {
				return err
			}
			fmt.Fprintln(u.Out, "Watching local files for changes...")
		}
	}
}

func (u *upCmd) doUp(environment string, cwd string, vals []byte) (err error) {
	env := u.Manifest.Environments[environment]
	if env.BuildTarPath != "" && env.ChartTarPath != "" {
		buildTar, e := os.Open(env.BuildTarPath)
		if e != nil {
			return e
		}
		chartTar, e := os.Open(env.ChartTarPath)
		if e != nil {
			return e
		}
		err = u.Client.Up(env.Name, env.Namespace, u.Out, buildTar, chartTar, vals)
	} else {
		err = u.Client.UpFromDir(env.Name, env.Namespace, u.Out, cwd, vals)
	}

	// format error before returning
	if err != nil {
		err = fmt.Errorf("there was an error running 'draft up': %v", err)
	}
	return
}

func defaultDraftEnvironment() string {
	env := os.Getenv(environmentEnvVar)
	if env == "" {
		env = manifest.DefaultEnvironmentName
	}
	return env
}

// removedFileInfo fake file info for
// ignore library only use IsDir() in negative pattern
type removedFileInfo string

func (n removedFileInfo) Name() string     { return string(n) }
func (removedFileInfo) Size() int64        { return 0 }
func (removedFileInfo) Mode() os.FileMode  { return 0 }
func (removedFileInfo) ModTime() time.Time { return time.Time{} }
func (removedFileInfo) IsDir() bool        { return false }
func (removedFileInfo) Sys() interface{}   { return nil }
