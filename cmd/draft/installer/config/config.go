package config

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// FromClientConfig reads a kubernetes client config, searching for information that may indicate
// this is a minikube/Azure Container Services/Google Container Engine cluster and return
// configuration optimized for that cloud, as well as the cloud provider name.
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
		baseDomain := os.Getenv("DRAFT_BASE_DOMAIN")
		if baseDomain == "" {
			baseDomain = "k8s.local"
		}

		// we imply that the user has installed the registry addon
		chartConfig.Raw = fmt.Sprintf("registry:\n  url: $(REGISTRY_SERVICE_HOST)\nbasedomain: %s\n", baseDomain)
		cloudProviderName = rawConfig.CurrentContext
	}

	return chartConfig, cloudProviderName, nil
}
