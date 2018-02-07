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
	Tunnels       []*kube.Tunnel
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
			var tt []*kube.Tunnel

			// iterate through all ports of the contaier and create tunnels
			for _, p := range c.Ports {
				t := kube.NewTunnel(clientset.CoreV1().RESTClient(), clientConfig, a.Namespace, pod.Name, int(p.ContainerPort))
				tt = append(tt, t)
			}
			cc = append(cc, &ContainerConnection{
				ContainerName: c.Name,
				Tunnels:       tt,
			})
		}

		return &Connection{
			ContainerConnections: cc,
			PodName:              pod.Name,
			Clientset:            clientset,
		}, nil
	}
	var tt []*kube.Tunnel

	// a container was specified - return tunnel to specified container
	ports, err := getTargetContainerPorts(pod.Spec.Containers, containerName)
	if err != nil {
		return nil, err
	}

	// iterate through all ports of the container and create tunnels
	for _, p := range ports {
		t := kube.NewTunnel(clientset.CoreV1().RESTClient(), clientConfig, a.Namespace, pod.Name, p)
		tt = append(tt, t)
	}

	cc = append(cc, &ContainerConnection{
		ContainerName: containerName,
		Tunnels:       tt,
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

func getTargetContainerPorts(containers []v1.Container, targetContainer string) ([]int, error) {
	var ports []int
	containerFound := false

	for _, c := range containers {

		if c.Name == targetContainer && !containerFound {
			containerFound = true
			for _, p := range c.Ports {
				ports = append(ports, int(p.ContainerPort))
			}
		}
	}

	if containerFound == false {
		return nil, fmt.Errorf("container '%s' not found", targetContainer)
	}

	return ports, nil
}
