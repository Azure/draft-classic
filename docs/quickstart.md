# Quickstart Guide

This guide covers how you can quickly get started using Draft

## Prerequisites

The following prerequisites are required for successful use of Draft.

1. A Kubernetes cluster and the `kubectl` CLI tool
2. Installation of [Helm](https://github.com/Kubernetes/helm) on your Kubernetes cluster
3. Installing and configuring Draft on your laptop

### Install Kubernetes or have access to a cluster
- If you don't already have a running Kubernetes cluster, consider downloading and installing [Minikube](https://github.com/kubernetes/minikube) which can help you get a Kubernetes cluster running on your local machine
- You can download and configure the `kubectl` binary using the directions [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

### Install and configure Helm
- Helm is a package manager for Kubernetes. It allows you to install and manage packages of Kubernetes manifests, more commonly known as "charts", on your Kubernetes cluster.
- Download the Helm binary using [Homebrew](https://brew.sh/) via `brew install kubernetes-helm` or the [official releases page](https://github.com/kubernetes/helm/releases).
- Once you have Helm on your machine and a running Kubernetes cluster, run the following command:
```console
$ helm init
```

### Install and configure Draft

#### Standalone Binary

Download the latest release of Draft from the [latest releases page](https://github.com/Azure/draft/releases/latest), unpack the binary and place it somewhere on your $PATH.

For example, for the v0.14.1 release, this can be done via

```console
$ wget https://azuredraft.blob.core.windows.net/draft/draft-v0.14.1-linux-amd64.tar.gz
$ wget https://azuredraft.blob.core.windows.net/draft/draft-v0.14.1-linux-amd64.tar.gz.sha256
```

Make sure to verify the contents have not been tampered with:

```console
$ cat draft-v0.14.1-linux-amd64.tar.gz.sha256
$ shasum -a 256 draft-v0.14.1-linux-amd64.tar.gz
```

Then unpack it and place it on your $PATH:

```console
$ tar -xzvf draft-v0.14.1-linux-amd64.tar.gz
$ sudo mv linux-amd64/draft /usr/local/bin/draft
```

Test it's working with

```console
$ draft version
```

#### Homebrew

To install Draft on MacOS using [Homebrew](https://brew.sh/):

```console
$ brew install azure/draft/draft
```

#### Chocolatey

To install Draft on Windows using [Chocolatey](https://chocolatey.org/):

```console
$ choco install draft
```

IMPORTANT: this package is [currently being maintained by the community](https://chocolatey.org/packages/draft) and not by any of the core maintainers. Always make sure to verify the security and contents of any untrusted package from the internet you are not familiar with.

#### GoFish

To install Draft on Windows/MacOS/Linux using [GoFish](https://gofi.sh):

```console
$ gofish install draft
```

#### Configure Draft

Once you've installed Draft, set it up on your machine by running:

```console
$ draft init
$ eval $(minikube docker-env)
```

The `eval $(minikube docker-env)` command allows Draft to build images directly using Minikube's Docker daemon which lets you skip having to set up a remote/external container registry.

Congratulations! You're all set! Check out the [Getting Started](getting-started.md) page to see how to use Draft with a sample application.
