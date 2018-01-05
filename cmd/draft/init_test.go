package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	installerConfig "github.com/Azure/draft/cmd/draft/installer/config"
	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/Azure/draft/pkg/draft/pack/repo"
)

func TestInitClientOnly(t *testing.T) {
	resetEnvVars := unsetEnvVars()
	tempHome, _ := tempDir(t, "draft-init")
	os.Setenv(homeEnvVar, tempHome)
	defer func() {
		teardown()
		resetEnvVars()
	}()

	cmd := &initCmd{
		clientOnly: true,
		out:        ioutil.Discard,
		in:         os.Stdin,
		home:       draftpath.Home(tempHome),
	}

	if err := cmd.run(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	plugins, err := findPlugins(cmd.home.Plugins())
	if err != nil {
		t.Fatal(err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %v", len(plugins))
	}

	repos := repo.FindRepositories(cmd.home.Packs())
	if len(repos) != 1 {
		t.Errorf("Expected 1 pack repo, got %v", len(repos))
	}
}

type configureMinikubeTestCase struct {
	answer             string
	autoAccept         bool
	expectedAutoAccept bool
}

func TestConfigureMinikube(t *testing.T) {
	// test the value of autoaccept

	testCases := []struct {
		answer             string
		autoAccept         bool
		expectedAutoAccept bool
	}{
		{
			answer:             "",
			autoAccept:         true,
			expectedAutoAccept: true,
		},
		{
			answer:             "Y",
			autoAccept:         false,
			expectedAutoAccept: true,
		},
		{
			answer:             "y",
			autoAccept:         false,
			expectedAutoAccept: true,
		},
		{
			answer:             "n",
			autoAccept:         false,
			expectedAutoAccept: false,
		},
		{
			answer:             "N",
			autoAccept:         false,
			expectedAutoAccept: false,
		},
		{
			answer:             "s",
			autoAccept:         false,
			expectedAutoAccept: false,
		},
	}

	for _, tc := range testCases {
		cmd := &initCmd{
			out: ioutil.Discard,
			in:  bytes.NewBufferString(tc.answer + "\n"),
		}

		if err := cmd.configureMinikube(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cmd.autoAccept != tc.expectedAutoAccept {
			t.Errorf("Expected auto accept to be %v, got %v", tc.expectedAutoAccept, cmd.autoAccept)
		}

	}

}

func TestSetupIngressAndBasedomain(t *testing.T) {
	testCases := []struct {
		ingressEnabled     bool
		basedomain         string
		expectedIngress    bool
		expectedBasedomain string
	}{
		{
			ingressEnabled:     false,
			expectedIngress:    false,
			expectedBasedomain: "",
		},
		{
			ingressEnabled:     true,
			basedomain:         " basedomain ",
			expectedIngress:    true,
			expectedBasedomain: "basedomain",
		},
	}

	for _, tc := range testCases {
		cmd := &initCmd{
			out:            ioutil.Discard,
			in:             bytes.NewBufferString(tc.basedomain + "\n"),
			ingressEnabled: tc.ingressEnabled,
		}

		draftConfig := &installerConfig.DraftConfig{}
		if err := cmd.setupIngressAndBasedomain(draftConfig); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if draftConfig.Ingress != tc.expectedIngress {
			t.Errorf("expected ingress enabled to be %v, got %v", tc.expectedIngress, draftConfig.Ingress)
		}

		if draftConfig.Basedomain != tc.expectedBasedomain {
			t.Errorf("expected basedomain to be %v, got %v", tc.expectedBasedomain, draftConfig.Basedomain)
		}
	}
}

func TestSetupContainerRegistry(t *testing.T) {
	resetEnvVars := unsetEnvVars()
	tempHome, _ := tempDir(t, "draft-setup-registry")
	os.Setenv(homeEnvVar, tempHome)
	defer func() {
		teardown()
		resetEnvVars()
	}()

	cmd := &initCmd{
		out:            ioutil.Discard,
		in:             bytes.NewBufferString("some registry\nsome username\n"),
		home:           draftpath.Home(tempHome),
		passwordReader: mockSecureReader{},
	}

	draftConfig := &installerConfig.DraftConfig{}
	if err := cmd.setupContainerRegistry(draftConfig); err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if draftConfig.RegistryURL != "some registry" {
		t.Errorf("Expected registry url to be 'some registry', got %v", draftConfig.RegistryURL)
	}

	expectedRegistryAuth := "eyJ1c2VybmFtZSI6InNvbWUgdXNlcm5hbWUiLCJwYXNzd29yZCI6InNvbWUgcGFzc3dvcmQifQ=="
	if draftConfig.RegistryAuth != expectedRegistryAuth {
		t.Errorf("Expected registry auth to be %v, got %v",
			expectedRegistryAuth, draftConfig.RegistryAuth)
	}
}

type mockSecureReader struct{}

func (r mockSecureReader) readPassword() ([]byte, error) {
	return []byte("some password"), nil
}
