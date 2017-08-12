package pack

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/helm/pkg/chartutil"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/Azure/draft/pkg/osutil"
)

const appPythonPath = "testdata/app-python"

const appEmptydirPath = "testdata/app-emptydir"

const pythonScript = `#!/bin/bash

APP_DIR=$1

# Exit early if app is clearly not Python.
if [ ! -f $APP_DIR/requirements.txt ]; then
  echo "app is clearly not Python. Exiting."
  exit 1
fi
echo "Python"
`

const testDockerfile = `FROM nginx:latest
`

const testDetect = `#!/bin/sh
echo "Test"
`

func TestDetect(t *testing.T) {
	p := &Pack{
		DetectScript: []byte(pythonScript),
	}

	output, err := p.Detect(appPythonPath)
	if err != nil {
		t.Error("expected detect to pass")
	}
	if output != "Python" {
		t.Errorf("expected output == 'Python', got '%s'", output)
	}

	// check that the absolute path to the directory name is passed in as $1
	p.DetectScript = []byte(`#!/bin/bash
echo $1`)
	abspath, _ := filepath.Abs(appPythonPath)
	output, err = p.Detect(appPythonPath)
	if err != nil {
		t.Error("expected detect to pass")
	}
	if output != abspath {
		t.Errorf("expected output == '%s', got '%s'", abspath, output)
	}

	// test with a bad dir
	_, err = p.Detect("/dir/does/not/exist")
	if err == nil {
		t.Error("expected err when running detect with a dir that does not exist")
	}

	// test an application that should fail detection
	p.DetectScript = []byte(pythonScript)
	output, err = p.Detect(appEmptydirPath)
	expectedErrType := reflect.TypeOf(new(exec.ExitError))
	expectedOutput := "app is clearly not Python. Exiting."
	if reflect.TypeOf(err) != expectedErrType {
		t.Errorf("expected '%v', got '%v'. Error: %v", expectedErrType, reflect.TypeOf(err), err)
	}
	if output != expectedOutput {
		t.Errorf("expected '%s', got '%s'", expectedOutput, output)
	}

	// check that an empty DetectScript returns an ExitError
	p.DetectScript = []byte{}
	output, err = p.Detect(appPythonPath)
	expectedOutput = "detect script not implemented"
	if reflect.TypeOf(err) != expectedErrType {
		t.Errorf("expected '%v', got '%v'. Error: %v", expectedErrType, reflect.TypeOf(err), err)
	}
	if output != expectedOutput {
		t.Errorf("expected '%s', got '%s'", expectedOutput, output)
	}
}

func TestSaveDir(t *testing.T) {
	p := &Pack{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: "chart-for-nigel-thornberry",
			},
		},
		Dockerfile:   []byte(testDockerfile),
		DetectScript: []byte(testDetect),
	}
	dir, err := ioutil.TempDir("", "draft-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := p.SaveDir(dir, false); err != nil {
		t.Errorf("expected there to be no error when writing to %v, got %v", dir, err)
	}

	detectPath := filepath.Join(dir, DetectName)
	exists, err := osutil.Exists(detectPath)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected SaveDir(dir, false) to not write the detect script to the resultant directory")
	}

	dockerfilePath := filepath.Join(dir, "Dockerfile")
	exists, err = osutil.Exists(dockerfilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("expected SaveDir(dir, false) to write the Dockerfile to %s", dockerfilePath)
	}

	chartPath := filepath.Join(dir, "chart", "chart-for-nigel-thornberry")
	exists, err = osutil.Exists(chartPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("expected SaveDir(dir, false) to write the chart to %s", chartPath)
	}

	badChartYamlPath := filepath.Join(dir, "chart", "Chart.yaml")
	exists, err = osutil.Exists(badChartYamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Errorf("expected SaveDir(dir, false) to NOT write a Chart.yaml to %s", badChartYamlPath)
	}

	goodChartYamlPath := filepath.Join(dir, "chart", p.Chart.Metadata.Name, "Chart.yaml")
	exists, err = osutil.Exists(goodChartYamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("expected SaveDir(dir, false) to write a Chart.yaml to %s", goodChartYamlPath)
	}

	if _, err := chartutil.LoadDir(filepath.Join(dir, "chart", p.Chart.Metadata.Name)); err != nil {
		t.Errorf("expected chart/ to be loadable by helm, got %s", err)
	}
}

func TestSaveDirDockerfileExistsInAppDir(t *testing.T) {
	p := &Pack{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: "chart-for-nigel-thornberry",
			},
		},
		Dockerfile:   []byte(testDockerfile),
		DetectScript: []byte(testDetect),
	}
	dir, err := ioutil.TempDir("", "draft-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	expectedDockerfile := []byte("FROM draft")
	if err := ioutil.WriteFile(tmpfn, expectedDockerfile, 0644); err != nil {
		t.Fatal(err)
	}

	if err := p.SaveDir(dir, false); err != nil {
		t.Errorf("expected there to be no error when writing to %v, got %v", dir, err)
	}

	savedDockerfile, err := ioutil.ReadFile(tmpfn)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(savedDockerfile, expectedDockerfile) {
		t.Errorf("expected '%s', got '%s'", string(expectedDockerfile), string(savedDockerfile))
	}
}
