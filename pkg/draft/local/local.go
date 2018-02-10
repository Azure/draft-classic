package local

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/kube"

	"github.com/Azure/draft/pkg/draft/manifest"
	"github.com/Azure/draft/pkg/kube/podutil"
)

// DraftLabelKey is the label selector key on a pod that allows
//  us to identify which draft app a pod is associated with
const DraftLabelKey = "draft"

// App encapsulates information about an application to connect to
//
//  Name is the name of the application
//  Namespace is the Kubernetes namespace it is deployed in
//  Container is the name the name of the application container to connect to
type App struct {
	Name      string
	Namespace string
	Container string
}

// Connection encapsulated information to connect to an application
type Connection struct {
	ContainerConnections []*ContainerConnection
	PodName              string
	Clientset            kubernetes.Interface
}

// ContainerConnection encapsulates a connection to a container in a pod
type ContainerConnection struct {
	Tunnel        *kube.Tunnel
	ContainerName string
}

// DeployedApplication returns deployment information about the deployed instance
//  of the source code given a path to your draft.toml file and the name of the
//  draft environment
func DeployedApplication(draftTomlPath, draftEnvironment string) (*App, error) {
	var draftConfig manifest.Manifest
	if _, err := toml.DecodeFile(draftTomlPath, &draftConfig); err != nil {
		return nil, err
	}

	appConfig, found := draftConfig.Environments[draftEnvironment]
	if !found {
		return nil, fmt.Errorf("Environment %v not found", draftEnvironment)
	}

	return &App{Name: appConfig.Name, Namespace: appConfig.Namespace}, nil
}

// Connect tunnels to a Kubernetes pod running the application and returns the connection information
func (a *App) Connect(clientset kubernetes.Interface, clientConfig *restclient.Config, containerName string) (*Connection, error) {

	var cc []*ContainerConnection

	pod, err := getPod(a.Namespace, a.Name, clientset)
	if err != nil {
		return nil, err
	}

	// if no container was specified as flag, return tunnels to all containers in pod
	if containerName == "" {

		for _, c := range pod.Spec.Containers {

			port := int(c.Ports[0].ContainerPort)
			t := kube.NewTunnel(clientset.CoreV1().RESTClient(), clientConfig, a.Namespace, pod.Name, port)
			cc = append(cc, &ContainerConnection{
				ContainerName: c.Name,
				Tunnel:        t,
			})
		}

		return &Connection{
			ContainerConnections: cc,
			PodName:              pod.Name,
			Clientset:            clientset,
		}, nil
	}

	// a container was specified - return tunnel to specified container
	port, err := getTargetContainerPort(pod.Spec.Containers, containerName)
	if err != nil {
		return nil, err
	}

	t := kube.NewTunnel(clientset.CoreV1().RESTClient(), clientConfig, a.Namespace, pod.Name, port)
	cc = append(cc, &ContainerConnection{
		ContainerName: containerName,
		Tunnel:        t,
	})

	return &Connection{
		ContainerConnections: cc,
		PodName:              pod.Name,
		Clientset:            clientset,
	}, nil

}

// RequestLogStream returns a stream of the application pod's logs
func (c *Connection) RequestLogStream(namespace string, containerName string, logLines int64) (io.ReadCloser, error) {
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(c.PodName,
		&v1.PodLogOptions{
			Follow:    true,
			TailLines: &logLines,
			Container: containerName,
		})

	return req.Stream()

}

func getPod(namespace, label string, clientset kubernetes.Interface) (*v1.Pod, error) {
	options := metav1.ListOptions{LabelSelector: labels.Set{DraftLabelKey: label}.AsSelector().String()}
	pods, err := clientset.CoreV1().Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find ready pod")
	}
	for _, p := range pods.Items {
		if podutil.IsPodReady(&p) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("could not find a ready pod")
}

func getTargetContainerPort(containers []v1.Container, targetContainer string) (int, error) {
	var port int
	containerFound := false

	for _, c := range containers {

		if c.Name == targetContainer && !containerFound {
			containerFound = true
			port = int(c.Ports[0].ContainerPort)
		}
	}

	if containerFound == false {
		return 0, fmt.Errorf("container '%s' not found", targetContainer)
	}

	return port, nil
}
