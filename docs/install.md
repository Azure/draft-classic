# Install Guide

Get started with Draft in three easy steps:

1. Install CLI tools for Helm, Kubectl, [Minikube][] and Draft
2. Boot Minikube and install Draft
3. Deploy your first application

## Dependencies

In order to get started, you will need to fetch the following:

- the latest release of minikube
- the latest release of kubectl
- the latest release of Helm

All of the dependencies can be installed by the following:

```
$ brew cask install minikube
```

Afterwards, fetch [the latest release of Draft](https://github.com/Azure/draft/releases).

Installing Draft via Homebrew can be done using

```
$ brew tap azure/draft
$ brew install draft
```

Canary releases of the Draft client can be found at the following links:

 - [Linux amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)
 - [Windows amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)

It can also be installed with

```
$ brew install draft-canary
```

Alternative downloads:

- [Linux ARM](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-arm.tar.gz)
- [Linux x86](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-386.tar.gz)

Unpack the Draft binary and add it to your PATH.

## Boot Minikube

At this point, you can boot up minikube!

```
$ minikube start
Starting local Kubernetes v1.7.3 cluster...
Starting VM...
oving files into cluster...
Setting up certs...
Starting cluster components...
Connecting to cluster...
Setting up kubeconfig...
Kubectl is now configured to use the cluster.
```

Now that the cluster is up and ready, minikube automatically configures kubectl, the command line tool for Kubernetes, on your machine with the appropriate authentication and endpoint information.

```
$ kubectl cluster-info
Kubernetes master is running at https://192.168.99.100:8443

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

## Enable Minikube Add-ons

Now that we have minikube running, we can go ahead and enable the `registry` add-on. The registry add-on is used to store the built docker container within the cluster.

You can enable the add-on with

```console
$ minikube addons enable registry
```

## Install Helm

Install Helm, a Kubernetes Package Manager, in your cluster. Helm manages the lifecycle of an application in Kubernetes, and it is also how Draft deploys an application to Kubernetes.

Installing Helm and setting it up is quite simple:

    $ helm init

Wait for Helm to come up and be in a `Ready` state. You can use `kubectl -n kube-system get deploy tiller-deploy --watch` to wait for tiller to come up.

## Install Draft

Now that all the dependencies are set up, we can set up Draft by running this command:

    $ draft init --auto-accept

Draft will read your local kube configuration and notice that it is pointing at minikube. It will then install Draftd (the Draft server) communicating with the installed registry add-on and Tiller (Helm server) instance.

## Take Draft for a Spin

Once you've completed the above steps, you're ready to climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!


[Getting Started]: getting-started.md
[minikube]: https://github.com/kubernetes/minikube
