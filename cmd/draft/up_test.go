package main

import (
	"path"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/Azure/draft/pkg/draft/manifest"
)

func TestUpVals(t *testing.T) {
	var env = new(manifest.Environment)
	appPath := path.Join("testdata", "create", "generated", "simple-go")
	if _, err := vals(env, appPath); err != nil {
		t.Errorf("expected no error with good data, got %v", err)
	}

	env.Values = []string{
		"replicaCount=3",
	}

	vals, err := vals(env, appPath)
	if err != nil {
		t.Errorf("expected no error with good data, got %v", err)
	}

	var parsedVals = make(map[string]interface{})
	if err := yaml.Unmarshal(vals, &parsedVals); err != nil {
		t.Error(err)
	}
	if parsedVals["replicaCount"] != 3 {
		t.Errorf("expected parsedVals['replicaCount'] = 3, got %d", parsedVals["replicaCount"])
	}
}
