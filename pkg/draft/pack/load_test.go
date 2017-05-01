package pack

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

	if _, err := FromDir("dir-does-not-exist"); err == nil {
		t.Errorf("expected err to be non-nil when path does not exist")
	}

	// post-cleanup: switch back to the cwd so other tests continue to work.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "draft-pack-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// remove the dir from under our feet to force filepath.Abs to fail
	os.RemoveAll(dir)
	if _, err := FromDir("."); err == nil {
		t.Errorf("expected err to be non-nil when filepath.Abs(\".\") should fail")
	}

	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("CI") != "" {
		t.Skip("skipping file permission mode tests on CI servers")
	}

	// re-use the dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if _, err := Create(path.Base(dir), path.Dir(dir), fooPackFiles()); err != nil {
		t.Fatal(err)
	}
	// load a pack with an un-readable Dockerfile (file perms 0000)
	if err := os.Chmod(filepath.Join(dir, DockerfileName), 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := FromDir(dir); err == nil {
		t.Errorf("expected err to be non-nil when reading the Dockerfile")
	}

	// revert file perms for the Dockerfile in prep for the detect script
	if err := os.Chmod(filepath.Join(dir, DockerfileName), 0644); err != nil {
		t.Fatal(err)
	}

	// make the detect script available but unreadable
	if err := os.Chmod(filepath.Join(dir, DetectName), 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := FromDir(dir); err == nil {
		t.Errorf("expected err to be non-nil when reading the detect script")
	}
}

func TestFromDirNoDetect(t *testing.T) {
	pack, err := FromDir("testdata/pack-python-no-detect")
	if err != nil {
		t.Fatalf("could not load python pack: %v", err)
	}
	if pack.Chart == nil {
		t.Errorf("expected chart to be non-nil")
	}

	if string(pack.Dockerfile) != expectedDockerfile {
		t.Errorf("expected dockerfile == expected, got '%v'", pack.Dockerfile)
	}

	if pack.DetectScript != nil {
		t.Errorf("expected detect to be nil, got '%v'", pack.DetectScript)
	}
}
