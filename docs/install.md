# Install Guide

Get started with Draft in three easy steps:

1. Install CLI tools for Helm, Kubectl, [Minikube][] and Draft
2. Boot Minikube and install Draft
3. Deploy your first application

## Dependencies

In order to get started, you will need to fetch the following:

- [the latest release of minikube](https://github.com/kubernetes/minikube/releases)
- [the latest release of kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [the latest release of Helm](https://github.com/kubernetes/helm/releases)
- [the latest release of Draft](https://github.com/Azure/draft/releases)

Canary releases of the Draft client can be found at the following links:

 - [Linux amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)
 - [Windows amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)

Alternative downloads:

- [Linux ARM](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-arm.tar.gz)
- [Linux x86](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-386.tar.gz)

Unpack the Draft binary and add it to your PATH.

## Enable Minikube Add-ons

Now that we have minikube installed, we can go ahead and enable the `registry` and `ingress`
add-ons.

The ingress add-on is used to allow inbound connections to reach the application.

The registry add-on is used to store the built docker container within the cluster.

You can enable the add-ons with

```
$ minikube addons enable ingress
$ minikube addons enable registry
```

## Boot Minikube

At this point, you can boot up minikube!

```
$ minikube start
Starting local Kubernetes v1.6.4 cluster...
Starting VM...
oving files into cluster...
Setting up certs...
Starting cluster components...
Connecting to cluster...
Setting up kubeconfig...
Kubectl is now configured to use the cluster.
```

Now that the cluster is up and ready, minikube automatically configures kubectl on your machine with
the appropriate authentication and endpoint information.

```
$ kubectl cluster-info
Kubernetes master is running at https://192.168.99.100:8443

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

## Install Helm

Once the cluster is ready, you will need to install Helm. Helm is a Kubernetes Package Manager and
is how Draft deploys an application to Kubernetes.

Installing Helm is quite simple:

```
$ helm init
```

Wait for Helm to come up and be in a `Ready` state. You can use `kubectl -n kube-system get deploy tiller-deploy --watch`
to wait for tiller to come up.

## Install Draft

Now that everything else is set up, we can now install Draft.

```
$ draft init
```

Follow through the prompts. Draft will read your local kube configuration and notice that it is
pointing at minikube. It will then install Draftd (the Draft server) communicating with the
installed registry add-on, ingress controller and Tiller (Helm server) instances.

## Configure Ingress Routes

Draft uses a wildcard domain to make accessing draft-created applications easier. To do so, it
specifies a custom host in the ingress from which tells the backing load balancer to route requests
based on the Host header.

When Draft was installed on Minikube, a base domain of `k8s.local` was used. To use this domain, you
can edit your `/etc/hosts` file to point to the ingressed out application domain to your cluster.

The following snippet would allow you to access an application:

```
$ sudo echo $(minikube ip) appname.k8s.local >> /etc/hosts
```

Unfortunately, `/etc/hosts` does not handle wildcard routes so each application deployed will need
to result in a new route in `/etc/hosts`. Others have worked around this by using other more
sophisticated tools like [dnsmasq][].

To use wildcard domains with dnsmasq, add a new rule in `dnsmasq.conf`:

```
$ sudo echo "address=/k8s.local/$(minikube ip)" >> dnsmasq.conf
```

See the [Ingress Guide][] for a more detailed setup.

## Take Draft for a Spin

Once you've completed the above steps, you're ready to climb aboard and explore the
[Getting Started Guide][Getting Started] - you'll soon be sailing!


[dnsmasq]: https://wiki.archlinux.org/index.php/dnsmasq
[Getting Started]: getting-started.md
[Ingress Guide]: ingress.md
[minikube]: https://github.com/kubernetes/minikube
