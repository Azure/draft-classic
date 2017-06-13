## Dependencies

- Draft will need a Kubernetes cluster to deploy your app.
  [Minikube][minikube], Azure Container Services and Google Container Engine
  are a few examples that will work with Draft, but any Kubernetes cluster will do.
- Draft expects [Helm](https://github.com/kubernetes/helm) to be installed on your Kubernetes cluster. Download [`helm` v2.4.x](https://github.com/kubernetes/helm/releases) and
do a `helm init` first, as described in [Installing Helm](https://github.com/kubernetes/helm/blob/master/docs/install.md).
- Draft needs to push images to a Docker registry, so you'll need to configure Draft with your Docker registry credentials. If you don't already have one, you can create a Docker registry for free on either [Docker Hub](https://hub.docker.com/) or [Quay.io](https://quay.io).
- An ingress controller installed within your Kubernetes cluster with a wildcard domain pointing to it. Review the [Ingress Guide][Ingress Guide] for more information about what Draft expects and how to set up an ingress controller.

## Install Draft

Because Draft is currently experimental, there is no stable release out yet and users are expected
to be using the latest build of Draft for testing. Canary releases of the Draft client can be found
at the following links:

 - [Linux amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)
 - Windows amd64 [coming soon!](https://github.com/Azure/draft/issues/61)

Unpack the Draft binary and add it to your PATH.

## Configure Draft

To install the server-side of Draft, use `draft init` with your ingress' `basedomain` and credentials to let Draft communicate with a Docker registry by using the following command:

```
$ draft init --set registry.url=changeme,registry.org=changeme,registry.authtoken=changeme,basedomain=changeme
```

* registry.url: Docker Registry Server URL. e.g. Azure Container Registry -> xxxx.azurecr.io, DockerHub -> docker.io
* basedomain: Using a domain that you manage. e.g. `draft.example.com` or use publicly available wildcard dns from [xip.io](https://xip.io). For minikube, as a result, basedomain could be `basedomain=$(minikube ip).xip.io`


The auth token field follows the format of Docker's X-Registry-Auth header.
For credential-based logins such as Azure Container Registry, Docker Hub and Quay, use:

```
$ echo '{"username":"jdoe","password":"secret","email":"jdoe@acme.com"}' | base64
```

For token-based logins such as Google Container Registry and Amazon ECR, use:

```
$ echo '{"registrytoken":"9cbaf023786cd7"}' | base64
```

## Take Draft for a Spin

Once you've completed the above steps, you're ready to climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!

[Ingress Guide]: ingress.md
[Getting Started]: getting-started.md
[minikube]: https://github.com/kubernetes/minikube
