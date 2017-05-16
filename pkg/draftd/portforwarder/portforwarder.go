package portforwarder

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"
)

const (
	// DraftNamespace is the Kubernetes namespace in which the Draft pod runs.
	DraftNamespace string = "kube-system"
)

// New returns a tunnel to the Draft pod.
func New(clientset *kubernetes.Clientset, config *restclient.Config) (*kube.Tunnel, error) {
	podName, err := getDraftPodName(clientset)
	if err != nil {
		return nil, err
	}
	const draftPort = 44135
	t := tunnel.New(clientset.CoreV1().RESTClient(), config, DraftNamespace, podName, draftPort)
	return t, t.ForwardPort()
}

func getDraftPodName(clientset *kubernetes.Clientset) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "draft", "name": "draftd"}.AsSelector()
	pod, err := getFirstRunningPod(clientset, selector)
	if err != nil {
		return "", err
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(clientset *kubernetes.Clientset, selector labels.Selector) (*api.Pod, error) {
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pods, err := clientset.CoreV1().Pods(DraftNamespace).List(options)
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
