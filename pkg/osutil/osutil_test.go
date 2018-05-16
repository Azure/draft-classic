package osutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	file, err := ioutil.TempFile("", "osutil")
	if err != nil {
		t.Fatal(err)
	}

	exists, err := Exists(file.Name())
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that exists, got %v", err)
	}
	if !exists {
		t.Error("expected tempfile to exist")
	}
	os.Remove(file.Name())
	exists, err = Exists(file.Name())
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that does not exist, got %v", err)
	}
	if exists {
		t.Error("expected tempfile to NOT exist")
	}
}

func TestSymlinkWithFallback(t *testing.T) {
	const (
		oldFileName = "foo.txt"
		newFileName = "bar.txt"
	)
	tmpDir, err := ioutil.TempDir("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldFileNamePath := filepath.Join(tmpDir, oldFileName)
	newFileNamePath := filepath.Join(tmpDir, newFileName)

	oldFile, err := os.Create(filepath.Join(tmpDir, oldFileName))
	if err != nil {
		t.Fatal(err)
	}
	oldFile.Close()

	if err := SymlinkWithFallback(oldFileNamePath, newFileNamePath); err != nil {
		t.Errorf("expected no error when calling SymlinkWithFallback() on a file that exists, got %v", err)
	}
}
