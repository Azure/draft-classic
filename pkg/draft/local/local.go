package local

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"

	"github.com/Azure/draft/pkg/draft/manifest"
)

const DraftLabelKey = "draft"

type app struct {
	Name      string
	Namespace string
}

type connection struct {
	Tunnel    *kube.Tunnel
	PodName   string
	Clientset kubernetes.Interface
}

// DeployedApplication returns deployment information about the deployed instance
//  of the source code given a path to your draft.toml file and the name of the
//  draft environment
func DeployedApplication(draftTomlPath, draftEnvironment string) (*app, error) {
	var draftConfig manifest.Manifest
	if _, err := toml.DecodeFile(draftTomlPath, &draftConfig); err != nil {
		return nil, err
	}
	appConfig := draftConfig.Environments[draftEnvironment]

	return &app{Name: appConfig.Name, Namespace: appConfig.Namespace}, nil
}

// Connect creates a local tunnel to a Kubernetes pod running the application and returns the connection information
func (a *app) Connect(clientset kubernetes.Interface, clientConfig *restclient.Config) (*connection, error) {
	tunnel, podName, err := a.NewTunnel(clientset, clientConfig)

	if err != nil {
		return nil, err
	}

	return &connection{
		Tunnel:    tunnel,
		PodName:   podName,
		Clientset: clientset,
	}, nil
}

// NewTunnel creates and returns a tunnel to a Kubernetes pod running the application
func (a *app) NewTunnel(clientset kubernetes.Interface, config *restclient.Config) (*kube.Tunnel, string, error) {
	podName, containers, err := getAppPodNameAndContainers(a.Namespace, a.Name, clientset)
	if err != nil {
		return nil, podName, err
	}

	// purposely handles first container and its first container port
	//  at the moment for MVP and should be changed later
	var port int
	for _, container := range containers {
		port = int(container.Ports[0].ContainerPort)
	}

	t := kube.NewTunnel(clientset.CoreV1().RESTClient(), config, a.Namespace, podName, port)
	if err != nil {
		return nil, podName, err
	}

	return t, podName, t.ForwardPort()
}

// RequestLogStream returns a stream of the application pod's logs
func (c *connection) RequestLogStream(app *app) (io.ReadCloser, error) {
	var lines int64 = 5

	req := c.Clientset.CoreV1().Pods(app.Namespace).GetLogs(c.PodName,
		&v1.PodLogOptions{
			Follow:    true,
			TailLines: &lines,
		})

	return req.Stream()
}

func getAppPodNameAndContainers(namespace, labelVal string, clientset kubernetes.Interface) (string, []v1.Container, error) {
	selector := labels.Set{DraftLabelKey: labelVal}.AsSelector()
	pod, err := getFirstRunningPod(clientset, selector, namespace)
	if err != nil {
		return "", nil, err
	}
	return pod.ObjectMeta.GetName(), pod.Spec.Containers, nil
}

func getFirstRunningPod(clientset kubernetes.Interface, selector labels.Selector, namespace string) (*v1.Pod, error) {
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pods, err := clientset.CoreV1().Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find ready pod")
	}
	for _, p := range pods.Items {
		if v1.IsPodReady(&p) {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("could not find a ready pod")
}
