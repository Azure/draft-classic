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

	"github.com/deis/prow/pkg/prow"
	"github.com/deis/prow/pkg/prow/pack"
	"github.com/deis/prow/pkg/strvals"
)

const upDesc = `
This command archives the current directory into a tar archive and uploads it to
the prow server.

Adding the "--watch" flag makes prow automatically archive and upload whenever
local files are saved. Prow delays a couple seconds to ensure that changes have
stopped before uploading, but that can be altered by the "--watch-delay" flag.
`

type upCmd struct {
	appName           string
	client            *prow.Client
	namespace         string
	out               io.Writer
	buildTarPath      string
	chartTarPath      string
	values            []string
	rawValueFilePaths []string
	wait              bool
	watch             bool
	watchDelay        int
}

func newUpCmd(out io.Writer) *cobra.Command {
	up := &upCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:     "up",
		Short:   "upload the current directory to the prow server for deployment",
		Long:    upDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			up.client = ensureProwClient(up.client)
			return up.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&up.appName, "app", "a", "", "name of the helm release. By default this is the basename of the current working directory")
	f.StringVarP(&up.namespace, "namespace", "n", "default", "kubernetes namespace to install the chart")
	f.StringVar(&up.buildTarPath, "build-tar", "", "path to a gzipped build tarball. --chart-tar must also be set.")
	f.StringVar(&up.chartTarPath, "chart-tar", "", "path to a gzipped chart tarball. --build-tar must also be set.")
	f.StringArrayVar(&up.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVarP(&up.rawValueFilePaths, "values", "f", []string{}, "specify prowd values in a YAML file (can specify multiple)")
	f.BoolVarP(&up.wait, "wait", "", false, "specifies whether or not to wait for all resources to be ready")
	f.BoolVarP(&up.watch, "watch", "w", false, "whether to deploy the app automatically when local files change")
	f.IntVarP(&up.watchDelay, "watch-delay", "", 2, "wait for local file changes to have stopped for this many seconds before deploying")

	return cmd
}

func (u *upCmd) vals(cwd string) ([]byte, error) {
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
	for _, filePath := range u.rawValueFilePaths {
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
	for _, value := range u.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

func (u *upCmd) run() (err error) {
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	if u.appName == "" {
		u.appName = path.Base(cwd)
	}
	u.client.OptionWait = u.wait

	rawVals, err := u.vals(cwd)
	if err != nil {
		return err
	}

	if err = u.doUp(cwd, rawVals); err != nil {
		return err
	}

	// if `--watch=false`, return now
	if !u.watch {
		return nil
	}

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
	delay := time.Duration(u.watchDelay) * time.Second

	for {
		select {
		case evt := <-ch:
			log.Debugf("Event %s", evt)
			// reset the timer when files have changed
			timer.Reset(delay)
		case <-timer.C:
			if err = u.doUp(cwd, rawVals); err != nil {
				return err
			}
			fmt.Fprintln(u.out, "Watching local files for changes...")
		}
	}
}

func (u *upCmd) doUp(cwd string, vals []byte) (err error) {
	if u.buildTarPath != "" && u.chartTarPath != "" {
		buildTar, e := os.Open(u.buildTarPath)
		if e != nil {
			return e
		}
		chartTar, e := os.Open(u.chartTarPath)
		if e != nil {
			return e
		}
		err = u.client.Up(u.appName, u.namespace, u.out, buildTar, chartTar, vals)
	} else {
		err = u.client.UpFromDir(u.appName, u.namespace, u.out, cwd, vals)
	}

	// format error before returning
	if err != nil {
		err = fmt.Errorf("there was an error running 'prow up': %v", err)
	}
	return
}
