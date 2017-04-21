package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/rjeczalik/notify"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/strvals"

	"github.com/deis/draft/pkg/draft"
	"github.com/deis/draft/pkg/draft/pack"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to
the draft server.

Adding the "--watch" flag makes draft automatically archive and upload whenever
local files are saved. Draft delays a couple seconds to ensure that changes have
stopped before uploading, but that can be altered by the "--watch-delay" flag.
`

const (
	environmentEnvVar        = "DRAFT_ENV"
	defaultEnvironmentName   = "development"
	defaultNamespace         = "default"
	defaultWatchDelaySeconds = 2
)

type upCmd struct {
	Client       *draft.Client           `json:"-"`
	Out          io.Writer               `json:"-"`
	Environments map[string]*Environment `json:"environments"`
}

// Environment represents the environment for a given app at build time
type Environment struct {
	AppName           string   `json:"name,omitempty"`
	BuildTarPath      string   `json:"build_tar,omitempty"`
	ChartTarPath      string   `json:"chart_tar,omitempty"`
	Namespace         string   `json:"namespace,omitempty"`
	Values            []string `json:"set,omitempty"`
	RawValueFilePaths []string `json:"values,omitempty"`
	Wait              bool     `json:"wait,omitempty"`
	Watch             bool     `json:"watch,omitempty"`
	WatchDelay        int      `json:"watch_delay,omitempty"`
}

func newUpCmd(out io.Writer) *cobra.Command {
	var (
		up = &upCmd{
			Out:          out,
			Environments: make(map[string]*Environment),
		}
		appName            string
		namespace          string
		buildTarPath       string
		chartTarPath       string
		values             []string
		rawValueFilePaths  []string
		runningEnvironment string
		wait               bool
		watch              bool
		watchDelay         int
	)

	cmd := &cobra.Command{
		Use:     "up",
		Short:   "upload the current directory to the draft server for deployment",
		Long:    upDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			up.Client = ensureDraftClient(up.Client)
			up.Environments[runningEnvironment] = &Environment{
				AppName:           appName,
				BuildTarPath:      buildTarPath,
				ChartTarPath:      chartTarPath,
				Namespace:         namespace,
				Values:            values,
				RawValueFilePaths: rawValueFilePaths,
				Wait:              wait,
				Watch:             watch,
				WatchDelay:        watchDelay,
			}
			draftYaml, err := ioutil.ReadFile("draft.yaml")
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				if err = yaml.Unmarshal(draftYaml, up); err != nil {
					return fmt.Errorf("could not unmarshal draft.yaml: %v", err)
				}
			}
			return up.run(runningEnvironment)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&appName, "app", "a", "", "name of the helm release. By default this is the basename of the current working directory")
	f.StringVarP(&namespace, "namespace", "n", defaultNamespace, "kubernetes namespace to install the chart")
	f.StringVarP(&runningEnvironment, "environment", "e", defaultDraftEnvironment(), "the environment (development, staging, qa, etc) that draft will run under")
	f.StringVar(&buildTarPath, "build-tar", "", "path to a gzipped build tarball. --chart-tar must also be set")
	f.StringVar(&chartTarPath, "chart-tar", "", "path to a gzipped chart tarball. --build-tar must also be set")
	f.StringArrayVar(&values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVarP(&rawValueFilePaths, "values", "f", []string{}, "specify draftd values in a YAML file (can specify multiple)")
	f.BoolVarP(&wait, "wait", "", false, "specifies whether or not to wait for all resources to be ready")
	f.BoolVarP(&watch, "watch", "w", false, "whether to deploy the app automatically when local files change")
	f.IntVarP(&watchDelay, "watch-delay", "", defaultWatchDelaySeconds, "wait for local file changes to have stopped for this many seconds before deploying")

	return cmd
}

func (e *Environment) vals(cwd string) ([]byte, error) {
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

	// User specified a values files via -f/--values
	for _, filePath := range e.RawValueFilePaths {
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
	for _, value := range e.Values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

func (u *upCmd) run(environment string) (err error) {
	env := u.Environments[environment]
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	if env.AppName == "" {
		env.AppName = path.Base(cwd)
	}
	u.Client.OptionWait = env.Wait

	rawVals, err := env.vals(cwd)
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
	env := u.Environments[environment]
	if env.BuildTarPath != "" && env.ChartTarPath != "" {
		buildTar, e := os.Open(env.BuildTarPath)
		if e != nil {
			return e
		}
		chartTar, e := os.Open(env.ChartTarPath)
		if e != nil {
			return e
		}
		err = u.Client.Up(env.AppName, env.Namespace, u.Out, buildTar, chartTar, vals)
	} else {
		err = u.Client.UpFromDir(env.AppName, env.Namespace, u.Out, cwd, vals)
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
		env = defaultEnvironmentName
	}
	return env
}
