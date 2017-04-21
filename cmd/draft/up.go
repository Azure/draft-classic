package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/rjeczalik/notify"
	"github.com/spf13/cobra"

	"github.com/deis/draft/pkg/draft"
	"github.com/deis/draft/pkg/draft/manifest"
	"github.com/deis/draft/pkg/draft/pack"
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
)

type upCmd struct {
	Client   *draft.Client
	Out      io.Writer
	Manifest *manifest.Manifest
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
		Use:     "up",
		Short:   "upload the current directory to the draft server for deployment",
		Long:    upDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			up.Client = ensureDraftClient(up.Client)
			up.Manifest = manifest.New()
			draftToml, err := ioutil.ReadFile("draft.toml")
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
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
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

	// create a timer to enforce a "quiet period" before deploying the app
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	delay := time.Duration(env.WatchDelay) * time.Second

	for {
		select {
		case evt := <-ch:
			log.Debugf("Event %s", evt)
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
