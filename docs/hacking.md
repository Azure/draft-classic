# Hacking on Prow

This document is for developers who are interested in working directly on the Prow codebase. In
this guide, we walk you through the process of setting up a development environment that is
suitable for hacking on Prow.

## Prerequisites

In order to successfully compile and test Prow binaries and build Docker images of Prowd, the
following are required:

 - [git](https://git-scm.com/)
 - [Go](https://golang.org/) 1.7 or later, with support for compiling to `linux/amd64`
 - [glide](https://github.com/Masterminds/glide)
 - [docker](https://www.docker.com/)
 - a Kubernetes cluster. We recommend [minikube](https://github.com/kubernetes/minikube)
 - [helm](https://github.com/kubernetes/helm)
 - [upx](https://upx.github.io) OPTIONAL if you want to compress your binary for smaller Docker images.

In most cases, you should simply install according to the instructions. We'll cover the special
cases below.

### Configuring Go

If your local workstation does not support the `linux/amd64` target environment, you will have to
install Go from source with cross-compile support for that environment. This is because some of the
components are built on your local machine and then injected into a Docker container.

Homebrew users can just install with cross compiling support:

```
$ brew install go --with-cc-common
```

It is also straightforward to build Go from source:

```
$ sudo su
$ curl -sSL https://storage.googleapis.com/golang/go1.7.5.src.tar.gz | tar -C /usr/local -xz
$ cd /usr/local/go/src
$ # compile Go for our default platform first, then add cross-compile support
$ ./make.bash --no-clean
$ GOOS=linux GOARCH=amd64 ./make.bash --no-clean
```

Once you can compile to `linux/amd64`, you should be able to compile Prowd for preparation into a
Docker image.

## Fork the Repository

Once the prerequisites have been met, we can begin to work with Prow.

Begin at Github by forking Prow, then clone that fork locally. Since Prow is written in Go, the
best place to put it is under `$GOPATH/src/github.com/deis/`.

```
$ mkdir -p $GOPATH/src/github.com/deis
$ cd $GOPATH/src/github.com/deis
$ git clone git@github.com:<username>/prow.git
$ cd prow
```

If you are going to be issuing pull requests to the upstream repository from which you forked, we
suggest configuring git such that you can easily rebase your code to the upstream repository's
master branch. There are various strategies for doing this, but the
[most common](https://help.github.com/articles/fork-a-repo/) is to add an `upstream` remote:

```
$ git remote add upstream https://github.com/deis/prow.git
```

## Make Your Changes

With your development environment set up and the code you wish to work on forked and cloned, you
can begin making your changes.

To do that, run the following to get your environment ready:

```
$ make bootstrap  # runs `glide install`
$ make build      # compiles `prow` and `prowd` inside bin/
```

## Test Your Changes

Prow includes a comprehensive suite of tests. To run these tests, run `make test`.

If you wish to run a more extensive set of tests, run `make test-e2e`.

## Deploying Your Changes

Although writing and executing tests are critical to ensuring code quality, you will likely want to
deploy changes to a live environment, whether to make use of those changes or to test them further.

To facilitate deploying Prowd images containing your changes to your Kubernetes cluster, you will
need to make use of a Docker Registry. This is a location to where you can push your custom-built
images and from where your Kubernetes cluster can retrieve those same images.

In most cases, a local registry will not be accessible to your Kubernetes nodes. A public registry
such as [DockerHub][dh] or [quay.io][quay] will suffice.

To use DockerHub for this purpose, for instance:

```
$ export DOCKER_REGISTRY="docker.io"
$ export IMAGE_PREFIX=<your DockerHub username>
```

To use quay.io:

```
$ export DEIS_REGISTRY="quay.io"
$ export IMAGE_PREFIX=<your quay.io username>
```

After your Docker Registry is set up, you can deploy your images using

```
$ make docker-build docker-push
```

To install Prowd, edit `chart/values.yaml` and change the `registry` and `name` values to the image
you just deployed:

```
$ $EDITOR charts/values.yaml
```

Then, install the chart:

```
$ helm install ./chart --name prow --namespace prow
```

You should see a new Helm release available in `helm list`.

### Re-deploying Your Changes

Because Prow deploys Kubernetes applications and Prow is a Kubernetes application itself, you can
use Prow to deploy Prow. How neat is that?!

To build your changes and upload it to Prowd, run

```
$ make build docker-binary
$ prow up
--> Building Dockerfile
--> Pushing 127.0.0.1:5000/prow:6f3b53003dcbf43821aea43208fc51455674d00e
--> Deploying to Kubernetes
--> code:DEPLOYED
```

You should see a new release of Prow available and deployed with `helm list`.
