package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Azure/draft/pkg/draft/draftpath"
)

func TestParseConfig(t *testing.T) {
	testCases := []struct {
		configFile  string
		expectErr   bool
		pluginCount int
		repoCount   int
	}{
		{"", false, 0, 0},
		{filepath.Join("testdata", "init", "configFile.toml"), false, 1, 1},
		{filepath.Join("testdata", "init", "malformedConfigFile.toml"), true, 0, 0},
		{filepath.Join("testdata", "init", "missingConfigFile.toml"), true, 0, 0},
	}

	for _, tc := range testCases {
		resetEnvVars := unsetEnvVars()
		tempHome, teardown := tempDir(t, "draft-init")
		defer func() {
			teardown()
			resetEnvVars()
		}()

		cmd := &initCmd{
			home:       draftpath.Home(tempHome),
			out:        ioutil.Discard,
			configFile: tc.configFile,
		}

		plugins, repos, err := cmd.parseConfig()
		if err != nil && !tc.expectErr {
			t.Errorf("Not expecting error but got error: %v", err)
		}
		if len(plugins) != tc.pluginCount {
			t.Errorf("Expected %v plugins, got %#v", tc.pluginCount, len(plugins))
		}
		if len(repos) != tc.repoCount {
			t.Errorf("Expected %v pack repos, got %#v", tc.repoCount, len(repos))
		}
	}
}
