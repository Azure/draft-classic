package osutil

import (
	"io/ioutil"
	"os"
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
