package pack

import (
	"testing"
)

const expectedDockerfile = `FROM python:onbuild

CMD [ "python", "./hello.py" ]

EXPOSE 80
`

const expectedDetect = `#!/usr/bin/env bash

APP_DIR=$1

# Exit early if app is clearly not Python.
if [ ! -f $APP_DIR/requirements.txt ] && [ ! -f $APP_DIR/setup.py ] && [ ! -f $APP_DIR/Pipfile ]; then
  exit 1
fi

echo Python
`

func TestFromDir(t *testing.T) {
	pack, err := FromDir("testdata/pack-python")
	if err != nil {
		t.Fatalf("could not load python pack: %v", err)
	}
	if pack.Chart == nil {
		t.Errorf("expected chart to be non-nil")
	}

	if string(pack.Dockerfile) != expectedDockerfile {
		t.Errorf("expected dockerfile == expected, got '%v'", pack.Dockerfile)
	}

	if string(pack.DetectScript) != expectedDetect {
		t.Errorf("expected detect == expected, got '%v'", pack.DetectScript)
	}
}
