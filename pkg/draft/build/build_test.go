package build

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlushToDisk(t *testing.T) {
	var err error
	LogRoot, err = ioutil.TempDir("", "draftpkgbuild")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(LogRoot)

	b := New()
	b.BuildImgLogs.WriteString("building")
	b.PushImgLogs.WriteString("pushing")
	b.ReleaseLogs.WriteString("releasing")

	if err := b.FlushToDisk(); err != nil {
		t.Errorf("expected no error when flushing to disk: %v", err)
	}

	expectedFiles := map[string]string{
		BuildImgLogFilename: "building",
		PushImgLogFilename:  "pushing",
		ReleaseLogFilename:  "releasing",
	}

	for fname, expectedMessage := range expectedFiles {
		b, err := ioutil.ReadFile(filepath.Join(LogRoot, b.ID, fname))
		if err != nil {
			t.Errorf("could not read file %s: %v", fname, err)
		}
		if strings.Compare(string(b), expectedMessage) != 0 {
			t.Errorf("messages differ. Got: '%s' Expected: '%s'", string(b), expectedMessage)
		}
	}
}

func TestLoadFromDisk(t *testing.T) {
	var err error
	LogRoot, err = ioutil.TempDir("", "draftpkgbuild")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(LogRoot)

	b := New()

	buildPath := filepath.Join(LogRoot, b.ID)
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		t.Fatal(err)
	}

	expectedFiles := map[string]string{
		BuildImgLogFilename: "building",
		PushImgLogFilename:  "pushing",
		ReleaseLogFilename:  "releasing",
	}

	for fname, expectedMessage := range expectedFiles {
		if err := ioutil.WriteFile(filepath.Join(LogRoot, b.ID, fname), []byte(expectedMessage), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := b.LoadFromDisk(); err != nil {
		t.Errorf("could not load logs from disk: %v", err)
	}

	if b.BuildImgLogs.String() != "building" {
		t.Errorf("build image logs differ. Got: '%s' Expected: '%s'", b.BuildImgLogs.String(), "building")
	}

	if b.PushImgLogs.String() != "pushing" {
		t.Errorf("push image logs differ. Got: '%s' Expected: '%s'", b.PushImgLogs.String(), "pushing")
	}

	if b.ReleaseLogs.String() != "releasing" {
		t.Errorf("release logs differ. Got: '%s' Expected: '%s'", b.ReleaseLogs.String(), "releasing")
	}
}
