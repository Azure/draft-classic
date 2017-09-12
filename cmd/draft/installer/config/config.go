package config

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// FromClientConfig reads a kubernetes client config, searching for information that may indicate
// this is a minikube/Azure Container Services/Google Container Engine cluster and return
// configuration optimized for that cloud, as well as the cloud provider name.
// Currently only supports minikube
func FromClientConfig(config clientcmd.ClientConfig) (*chart.Config, string, error) {
	var (
		chartConfig       = new(chart.Config)
		cloudProviderName string
	)

	rawConfig, err := config.RawConfig()
	if err != nil {
		return nil, "", err
	}

	if rawConfig.CurrentContext == "minikube" {
		// we imply that the user has installed the registry addon
		chartConfig.Raw = "registry:\n  url: $(REGISTRY_SERVICE_HOST)\nbasedomain: k8s.local\n"
		cloudProviderName = rawConfig.CurrentContext
	}

	return chartConfig, cloudProviderName, nil
}
