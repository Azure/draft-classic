package pack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/Azure/draft/pkg/osutil"
)

const (
	// MetadataName is the name of the pack metadata
	MetadataName = "Pack.yaml"
	// ChartfileName is the default Chart file name.
	ChartfileName = "Chart.yaml"
	// ChartDir is the relative directory name for the packaged chart with a pack.
	ChartDir = "chart"
	// DetectName is the name of the detect script.
	DetectName = "detect"
	// DockerfileName is the name of the Dockerfile.
	DockerfileName = "Dockerfile"
	// ValuesfileName is the default values file name.
	ValuesfileName = "values.yaml"
	// IgnorefileName is the name of the Helm ignore file.
	IgnorefileName = ".helmignore"
	// DeploymentName is the name of the deployment file.
	DeploymentName = "deployment.yaml"
	// ServiceName is the name of the service file.
	ServiceName = "service.yaml"
	// IngressName is the name of the ingress file.
	IngressName = "ingress.yaml"
	// NotesName is the name of the NOTES.txt file.
	NotesName = "NOTES.txt"
	// HelpersName is the name of the helpers file.
	HelpersName = "_helpers.tpl"
	// TemplatesDir is the relative directory name for templates.
	TemplatesDir = "templates"
	// ChartsDir is the relative directory name for charts dependencies.
	ChartsDir = "charts"
	// HerokuLicenseName is the name of the Heroku License
	HerokuLicenseName = "NOTICE"
	// DockerignoreName is the name of the Docker ignore file
	DockerignoreName = ".dockerignore"
)

// Metadata defines the metadata for a Pack
type Metadata struct {
	// The name of the pack
	Name        string
	Version     string
	Description string
	Keywords    []string
}

// Pack defines a Draft Starter Pack.
type Pack struct {
	// The metadata for the Pack
	Metadata *Metadata
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
		// https://github.com/Azure/draft/blob/master/docs/packs.md#pack-detection
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

	bashPath := "/bin/bash"
	if runtime.GOOS == "windows" {
		// There is no common absolute path on Windows, ensure bash.exe is in the PATH.
		bashPath = "bash.exe"
	}

	cmd := exec.Command(bashPath, "-s", path)
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
	chartPath := filepath.Join(dest, ChartDir)
	if err := os.Mkdir(chartPath, 0755); err != nil {
		return fmt.Errorf("Could not create %s: %s", chartPath, err)
	}
	if err := chartutil.SaveDir(p.Chart, chartPath); err != nil {
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
