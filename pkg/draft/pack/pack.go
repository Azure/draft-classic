package pack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/deis/draft/pkg/osutil"
)

// Pack defines a Draft Starter Pack.
type Pack struct {
	// Chart is the Helm chart to be installed with the Pack.
	Chart *chart.Chart
	// Dockerfile is the pre-defined Dockerfile that will be installed with the Pack.
	Dockerfile []byte
	// DetectScript is a command that determines if the Pack is a candidate for an app. When
	// .Detect() is called on the Pack, the data here is piped as stdin to `/bin/bash -s`.
	DetectScript []byte
}

// Detect determines if this pack is viable for the given application in dir.
//
// returns a nil err if it is a viable pack. The string returned is the output of the
// detect script.
func (p *Pack) Detect(dir string) (string, error) {
	if len(p.DetectScript) == 0 {
		// If no detect script was implemented, we can assume it's a non-zero exit code.
		// https://github.com/deis/draft/blob/master/docs/packs.md#pack-detection
		return "detect script not implemented", &exec.ExitError{}
	}

	path, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if fi, err := os.Stat(path); err != nil {
		return "", err
	} else if !fi.IsDir() {
		return "", fmt.Errorf("no such directory %s", path)
	}

	cmd := exec.Command("/bin/bash", "-s", path)
	cmd.Stdin = bytes.NewBuffer(p.DetectScript)
	out, err := cmd.Output()
	if err != nil {
		return strings.TrimSpace(string(out)), err
	}
	return strings.TrimSpace(string(out)), nil
}

// SaveDir saves a pack as files in a directory.
func (p *Pack) SaveDir(dest string, includeDetectScript bool) error {
	// Create the chart directory
	chartName := p.Chart.Metadata.Name
	// HACK(bacongobbler): chartutil.SaveDir uses the chart name as the dirname. Because we want to
	// write it to chart/, we name the chart 'chart' and then re-save the Chart.yaml later with
	// chartutil.SaveChartfile().
	p.Chart.Metadata.Name = "chart"
	if err := chartutil.SaveDir(p.Chart, dest); err != nil {
		return err
	}
	p.Chart.Metadata.Name = chartName
	if err := chartutil.SaveChartfile(filepath.Join(dest, ChartDir, ChartfileName), p.Chart.Metadata); err != nil {
		return err
	}

	// save Dockerfile
	dockerfilePath := filepath.Join(dest, DockerfileName)
	exists, err := osutil.Exists(dockerfilePath)
	if err != nil {
		return err
	}
	if !exists {
		if err := ioutil.WriteFile(dockerfilePath, p.Dockerfile, 0644); err != nil {
			return err
		}
	}

	if includeDetectScript {
		// Save detect script
		detectPath := filepath.Join(dest, DetectName)
		if err := ioutil.WriteFile(detectPath, p.DetectScript, 0755); err != nil {
			return err
		}
	}
	return nil
}
