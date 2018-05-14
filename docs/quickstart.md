# Quickstart Guide

This guide covers how you can quickly get started using Draft

## Prerequisites

The following prerequisites are required for successful use of Draft.

1. A Kubernetes cluster and the `kubectl` CLI tool
2. Installation of Helm on your Kubernetes cluster
3. Installing and configuring Draft on your laptop

### Install Kubernetes or have access to a cluster
- If you don't already have a running Kubernetes cluster, consider downloading and installing [Minikube][] which can help you get a Kubernetes cluster running on your local machine
- You can download and configure the `kubectl` binary using the directions [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

### Install and configure Helm
- Helm is a package manager for Kubernetes. It allows you to install and manage packages of Kubernetes manifests, more commonly known as "charts", on your Kubernetes cluster.
- Download the Helm binary using [Homebrew](https://brew.sh/) via `brew install kubernetes-helm` or the [official releases page](https://github.com/kubernetes/helm/releases).
- Once you have Helm on your machine and a running Kubernetes cluster, run the following command:
```console
$ helm init
```

### Install and configure Draft
- Download the Draft binary using Homebrew via the commands below or from the [official releases page](https://github.com/Azure/draft/releases)
```console
$ brew tap azure/draft
$ brew install azure/draft/draft
```
- Set up Draft on your machine by running:
```console
$ draft init
$ eval $(minikube docker-env)
```

The `eval $(minikube docker-env)` command allows Draft to build images directly using Minikube's Docker daemon which lets you skip having to set up a remote/external container registry.

Congratulations! You're all set! Check out the [Getting Started](getting-started.md) page to see how to use Draft with a sample application.
