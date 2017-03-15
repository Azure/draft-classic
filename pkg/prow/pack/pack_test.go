package pack

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/helm/pkg/proto/hapi/chart"
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
		Dockerfile:   []byte(defaultDockerfile),
		DetectScript: []byte(defaultDetect),
	}
	dir, err := ioutil.TempDir("", "prow-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := p.SaveDir(dir, false); err != nil {
		t.Errorf("expected there to be no error when writing to %v, got %v", dir, err)
	}

	detectPath := filepath.Join(dir, DetectName)
	exists, err := exists(detectPath)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected SaveDir(dir, false) to not write the detect script to the resultant directory")
	}
}

func TestSaveDirDockerfileExists(t *testing.T) {
	p := &Pack{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: "chart-for-nigel-thornberry",
			},
		},
		Dockerfile:   []byte(defaultDockerfile),
		DetectScript: []byte(defaultDetect),
	}
	dir, err := ioutil.TempDir("", "prow-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	expectedDockerfile := []byte("FROM prow")
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
