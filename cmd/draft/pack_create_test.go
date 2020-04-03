package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestPackCreate(t *testing.T) {
	tDir, teardown := tempDir(t, "draft-pack-create")

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tDir); err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(pwd)
	defer teardown()

	packName := "example-pack-name"
	cmd := newPackCreateCmd(ioutil.Discard)
	if err := cmd.RunE(cmd, []string{packName}); err != nil {
		t.Errorf("Failed to run `pack create`: %s", err)
	}

	if fi, err := os.Stat(packName); err != nil {
		t.Fatalf("no pack directory: %s", err)
	} else if !fi.IsDir() {
		t.Fatalf("pack is not directory")
	}

	cmd = newPackCreateCmd(ioutil.Discard)
	if err := cmd.RunE(cmd, []string{packName}); err == nil {
		t.Error("Expected err, got none")
	}
}
