package portforwarder

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/labels"

	"k8s.io/helm/pkg/kube"
)

const (
	// ProwNamespace is the Kubernetes namespace in which the prow pod runs.
	ProwNamespace string = "kube-system"
)

// New returns a tunnel to the prow pod.
func New(context string) (*kube.Tunnel, error) {
	config, client, err := getKubeClient(context)
	if err != nil {
		return nil, err
	}

	podName, err := getProwPodName(client.Core())
	if err != nil {
		return nil, err
	}
	const prowPort = 44135
	t := kube.NewTunnel(client.Core().RESTClient(), config, ProwNamespace, podName, prowPort)
	return t, t.ForwardPort()
}

func getProwPodName(client internalversion.PodsGetter) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "prow", "name": "prowd"}.AsSelector()
	pod, err := getFirstRunningPod(client, selector)
	if err != nil {
		return "", err
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(client internalversion.PodsGetter, selector labels.Selector) (*api.Pod, error) {
	options := api.ListOptions{LabelSelector: selector}
	pods, err := client.Pods(ProwNamespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find prowd")
	}
	for _, p := range pods.Items {
		if api.IsPodReady(&p) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("could not find a ready prowd pod")
}

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*restclient.Config, *internalclientset.Clientset, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return config, client, nil
}
