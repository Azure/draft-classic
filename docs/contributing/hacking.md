# Hacking on Draft

This guide is for developers who want to improve Draft. These instructions will help you set up a
development environment for working on the Draft source code.

## Prerequisites

To compile and test Draft binaries and to build Docker images, you will need:

 - [docker][]
 - a [Docker Hub][] or [quay.io][quay] account
 - [git][]
 - [Go][] 1.7 or later, with support for compiling to `linux/amd64`
 - [glide][]
 - a [Kubernetes][] cluster. We recommend [minikube][]
 - [helm][]
 - [upx][] (optional) to compress binaries for a smaller Docker image

In most cases, install the prerequisite according to its instructions. See the next section
for a note about Go cross-compiling support.

### Configuring Go

Draft's binary executables are built on your machine and then copied into a Docker image. This
requires your Go compiler to support the `linux/amd64` target architecture. If you develop on a
machine that isn't AMD64 Linux, make sure that `go` has support for cross-compiling.

On macOS, a cross-compiling Go can be installed with [Homebrew][]:

```shell
$ brew install go --with-cc-common
```

It is also straightforward to build Go from source:

```shell
$ sudo su
$ curl -sSL https://storage.googleapis.com/golang/go1.7.5.src.tar.gz | tar -C /usr/local -xz
$ cd /usr/local/go/src
$ # compile Go for the default platform first, then add cross-compile support
$ ./make.bash --no-clean
$ GOOS=linux GOARCH=amd64 ./make.bash --no-clean
```

## Fork the Repository

Begin at Github by forking Draft, then clone your fork locally. Since Draft is a Go package, it
should be located at `$GOPATH/src/github.com/Azure/draft`.

```shell
$ mkdir -p $GOPATH/src/github.com/Azure
$ cd $GOPATH/src/github.com/Azure
$ git clone git@github.com:<username>/draft.git
$ cd draft
```

Add the conventional [upstream][] `git` remote in order to fetch changes from Draft's main master
branch and to create pull requests:

```shell
$ git remote add upstream https://github.com/Azure/draft.git
```

## Build Your Changes

With the prerequisites installed and your fork of Draft cloned, you can make changes to local Draft
source code.

Run `make` to build the `draft` and `draftd` binaries:

```shell
$ make bootstrap  # runs `glide install`
$ make build      # compiles `draft` and `draftd` inside bin/
```

## Test Your Changes

Draft includes a suite of tests. Run `make test` for basic unit tests or `make test-e2e` for more
comprehensive, end-to-end tests.

## Deploying Your Changes

To test interactively, you will likely want to deploy your changes to Draft on a Kubernetes cluster.
This requires a Docker registry where you can push your customized draftd images so Kubernetes can
pull them.

In most cases, a local Docker registry will not be accessible to your Kubernetes nodes. A public
registry such as [Docker Hub][] or [Quay][] will suffice.

To use DockerHub for draftd images:

```shell
$ export DOCKER_REGISTRY="docker.io"
$ export IMAGE_PREFIX=<your DockerHub username>
```

To use quay.io:

```shell
$ export DOCKER_REGISTRY="quay.io"
$ export IMAGE_PREFIX=<your quay.io username>
```

After your Docker registry is set up, you can deploy your images using:

```shell
$ make docker-build docker-push
```

Ensure that Helm's `tiller` server is running in the Kubernetes cluster:

```shell
$ helm init  # it may take a few seconds for tiller to install
$ helm version
Client: &version.Version{SemVer:"v2.2.0", GitCommit:"fc315ab59850ddd1b9b4959c89ef008fef5cdf89", GitTreeState:"clean"}
Server: &version.Version{SemVer:"v2.2.0", GitCommit:"fc315ab59850ddd1b9b4959c89ef008fef5cdf89", GitTreeState:"clean"}
```

To install draftd, edit `chart/values.yaml` and change the fields under `registry` to your
[Docker Hub][] or [quay.io][quay] account, and change the fields under `image` to the newly
deployed draftd image:

```
$ $EDITOR chart/values.yaml
```

Then, install the chart:

```shell
$ draft init -f chart/values.yaml
$ helm list  # check that Draft has a helm release
NAME 	REVISION	UPDATED                 	STATUS  	CHART      	NAMESPACE
draft	1       	Thu Feb 16 10:18:21 2017	DEPLOYED	draftd-0.1.0	kube-system
```

## Re-deploying Your Changes

Because Draft deploys Kubernetes applications and Draft is a Kubernetes application itself, you can
use Draft to deploy Draft. How neat is that?!

To build your changes and upload it to draftd, run

```shell
$ make build docker-binary
$ draft up
--> Building Dockerfile
--> Pushing docker.io/microsoft/draftd:6f3b53003dcbf43821aea43208fc51455674d00e
--> Deploying to Kubernetes
--> Status: DEPLOYED
--> Notes:
     Now you can deploy an app using Draft!

        $ cd my-app
        $ draft create
        $ draft up
        --> Building Dockerfile
        --> Pushing my-app:latest
        --> Deploying to Kubernetes
        --> Deployed!

That's it! You're now running your app in a Kubernetes cluster.
```

You should see a new release of Draft available and deployed with `helm list`.

## Cleaning Up

To remove the Draft chart and local binaries:

```shell
$ make clean unserve
rm bin/*
rm rootfs/bin/*
helm delete --purge draft
```


[docker]: https://www.docker.com/
[Docker Hub]: https://hub.docker.com/
[git]: https://git-scm.com/
[glide]: https://github.com/Masterminds/glide
[go]: https://golang.org/
[helm]: https://github.com/kubernetes/helm
[Homebrew]: https://brew.sh/
[Kubernetes]: https://github.com/kubernetes/kubernetes
[minikube]: https://github.com/kubernetes/minikube
[Quay]: https://quay.io/
[upstream]: https://help.github.com/articles/fork-a-repo/
[upx]: https://upx.github.io
