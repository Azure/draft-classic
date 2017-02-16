# Hacking on Prow

This guide is for developers who want to improve Prow. These instructions will help you set up a
development environment for working on the Prow source code.

## Prerequisites

To compile and test Prow binaries and to build Docker images, you will need:

 - [docker][]
 - [git][]
 - [Go][] 1.7 or later, with support for compiling to `linux/amd64`
 - [glide][]
 - a [Kubernetes][] cluster. We recommend [minikube][]
 - [helm][]
 - [upx][] (optional) to compress binaries for a smaller Docker image

In most cases, install the prerequisite according to its instructions. See the next section
for a note about Go cross-compiling support.

### Configuring Go

Prow's binary executables are built on your machine and then copied into a Docker image. This
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

Begin at Github by forking Prow, then clone your fork locally. Since Prow is a Go package, it
should be located at `$GOPATH/src/github.com/deis/prow`.

```shell
$ mkdir -p $GOPATH/src/github.com/deis
$ cd $GOPATH/src/github.com/deis
$ git clone git@github.com:<username>/prow.git
$ cd prow
```

Add the conventional [upstream][] `git` remote in order to fetch changes from Prow's main master
branch and to create pull requests:

```shell
$ git remote add upstream https://github.com/deis/prow.git
```

## Build Your Changes

With the prerequisites installed and your fork of Prow cloned, you can make changes to local Prow
source code.

Run `make` to build the `prow` and `prowd` binaries:

```shell
$ make bootstrap  # runs `glide install`
$ make build      # compiles `prow` and `prowd` inside bin/
```

## Test Your Changes

Prow includes a suite of tests. Run `make test` for basic unit tests or `make test-e2e` for more
comprehensive, end-to-end tests.

## Deploying Your Changes

To test interactively, you will likely want to deploy your changes to Prow on a Kubernetes cluster.
This requires a Docker registry where you can push your customized Prowd images so Kubernetes can
pull them.

In most cases, a local Docker registry will not be accessible to your Kubernetes nodes. A public
registry such as [Docker Hub][] or [Quay][] will suffice.

To use DockerHub for Prowd images:

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

Then, install the Prow chart:

```shell
$ make serve
$ helm list  # check that prow has a helm release
NAME 	REVISION	UPDATED                 	STATUS  	CHART      	NAMESPACE
prow	1       	Thu Feb 16 10:18:21 2017	DEPLOYED	prowd-0.1.0	prow
```

## Re-deploying Your Changes

Because Prow deploys Kubernetes applications and Prow is a Kubernetes application itself, you can
use Prow to deploy Prow. How neat is that?!

To build your changes and upload it to Prowd, run

```shell
$ make build docker-binary
$ prow up
--> Building Dockerfile
--> Pushing 127.0.0.1:5000/prow:6f3b53003dcbf43821aea43208fc51455674d00e
--> Deploying to Kubernetes
--> code:DEPLOYED
```

You should see a new release of Prow available and deployed with `helm list`.

## Cleaning Up

To remove the Prow chart and local binaries:

```shell
$ make clean
helm delete --purge prowd
rm bin/*
rm rootfs/bin/*
```


[docker]: https://www.docker.com/
[Docker Hub]: https://hub.docker.com/
[git]: https://git-scm.com/
[glide]: https://github.com/Masterminds/glide
[go]: https://golang.org/
[helm]: https://github.com/kubernetes/helm
[Homebrew]: https://brew.sh/
[Kubernetes]: https://github.com/kubernetes/minikube
[minikube]: https://github.com/kubernetes/minikube
[Quay]: https://quay.io/
[upstream]: https://help.github.com/articles/fork-a-repo/
[upx]: https://upx.github.io
