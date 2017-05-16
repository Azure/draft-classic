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
	// DraftNamespace is the Kubernetes namespace in which the Draft pod runs.
	DraftNamespace string = "kube-system"
)

// New returns a tunnel to the Draft pod.
func New(client *internalclientset.Clientset, config *restclient.Config) (*kube.Tunnel, error) {
	podName, err := getDraftPodName(client.Core())
	if err != nil {
		return nil, err
	}
	const draftPort = 44135
	t := kube.NewTunnel(client.Core().RESTClient(), config, DraftNamespace, podName, draftPort)
	return t, t.ForwardPort()
}

func getDraftPodName(client internalversion.PodsGetter) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "draft", "name": "draftd"}.AsSelector()
	pod, err := getFirstRunningPod(client, selector)
	if err != nil {
		return "", err
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(client internalversion.PodsGetter, selector labels.Selector) (*api.Pod, error) {
	options := api.ListOptions{LabelSelector: selector}
	pods, err := client.Pods(DraftNamespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find draftd")
	}
	for _, p := range pods.Items {
		if api.IsPodReady(&p) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("could not find a ready draftd pod")
}
