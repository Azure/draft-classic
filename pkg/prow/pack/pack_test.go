package pack

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
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
