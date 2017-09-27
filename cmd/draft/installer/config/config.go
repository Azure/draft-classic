package config

import (
	"k8s.io/client-go/tools/clientcmd"
)

type DraftConfig struct {
	Basedomain   string
	Image        string
	Ingress      bool
	RegistryURL  string
	RegistryAuth string
}

// FromClientConfig reads a kubernetes client config, searching for information that may indicate
// this is a minikube/Azure Container Services/Google Container Engine cluster and return
// configuration optimized for that cloud, as well as the cloud provider name.
// Currently only supports minikube
func FromClientConfig(config clientcmd.ClientConfig) (*DraftConfig, string, error) {
	var cloudProviderName string

	rawConfig, err := config.RawConfig()
	if err != nil {
		return nil, "", err
	}

	draftConfig := &DraftConfig{}

	if rawConfig.CurrentContext == "minikube" {
		// we imply that the user has installed the registry addon
		draftConfig.RegistryURL = "$(REGISTRY_SERVICE_HOST)"
		draftConfig.RegistryAuth = "e30K"
		draftConfig.Basedomain = "k8s.local"
		cloudProviderName = rawConfig.CurrentContext
	}

	return draftConfig, cloudProviderName, nil
}
