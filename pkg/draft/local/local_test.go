package local

import (
	"reflect"
	"testing"
)

func TestDeployedApplication(t *testing.T) {
	expectedApp := &App{
		Name:      "example-app",
		Namespace: "example-namespace",
	}

	app, err := DeployedApplication("../testdata/app/draft.toml", "development")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedApp, app) {
		t.Errorf("Expected %#v, got %#v", expectedApp, app)
	}
}
