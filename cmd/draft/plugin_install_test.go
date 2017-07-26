package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

type pluginTest struct {
	name   string
	plugin string
	path   string
	output string
	fail   bool
	flags  []string
}

func TestPluginInstallCmd(t *testing.T) {
	// set up draft home
	old := draftHome
	defer func() {
		draftHome = old
	}()

	dir, err := ioutil.TempDir("", "draft_home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	draftHome = dir
	home := draftpath.Home(draftHome)

	if err := os.Mkdir(home.Plugins(), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []pluginTest{
		{
			name:   "install plugin",
			plugin: "echo",
			path:   "testdata/plugins/echo",
			output: "Installed plugin: echo",
			fail:   false,
		},
		{
			name:   "error installing nonexistent plugin",
			plugin: "dummy",
			path:   "testdata/plugins/dummy",
			output: "",
			fail:   true,
		},
	}

	buf := bytes.NewBuffer(nil)
	for _, tt := range tests {
		cmd := newPluginInstallCmd(buf)

		if err := cmd.PreRunE(cmd, []string{tt.path}); err != nil {
			t.Errorf("%q reported error: %s", tt.name, err)
		}

		if err := cmd.RunE(cmd, []string{tt.path}); err != nil && !tt.fail {
			t.Errorf("%q reported error: %s", tt.name, err)
		}

		if !tt.fail {
			result := buf.String()
			if !strings.Contains(result, tt.output) {
				t.Errorf("Expected %v, got %v", tt.output, result)
			}

			if _, err = os.Stat(filepath.Join(home.Plugins(), tt.plugin)); err != nil && os.IsNotExist(err) {
				t.Errorf("Installed plugin not found: %v", err)
			}

		}

		buf.Reset()
	}

	cmd := newPluginInstallCmd(buf)
	if err := cmd.PreRunE(cmd, []string{"arg1", "extra arg"}); err == nil {
		t.Error("Expected failure due to incorrect number of arguments for plugin install command")
	}

}
