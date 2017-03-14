# Prow: Streamlined Kubernetes Development

[![Build Status](http://drone.champagne.deis.com/api/badges/deis/prow/status.svg)](http://drone.champagne.deis.com/deis/prow)

_NOTE: Prow is experimental and does not have a stable release yet._

Prow handles the heavy lifting involved in taking source code and deploying it to Kubernetes:

- Builds a container image from application source code
- Pushes the image to a registry
- Packages a [Helm][] chart from application source code
- Installs the chart to Kubernetes, deploying the application

## Usage

### Install Prow

Because Prow is currently experimental, there is no stable release out yet and users are expected
to be using the latest build of Prow for testing. Canary releases of the Prow client can be found
at the following links:

 - [Linux amd64](https://s3-us-west-2.amazonaws.com/deis-prow/prow-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://s3-us-west-2.amazonaws.com/deis-prow/prow-canary-darwin-amd64.tar.gz)
 - [Windows amd64](https://s3-us-west-2.amazonaws.com/deis-prow/prow-canary-windows-amd64.tar.gz)

Unpack the Prow binary and add it to your PATH and you are good to go!

To install the server-side of Prow, use `prow init` with your credentials to let Prow communicate
with a Docker registry:

```
$ prow init --set registry.url=docker.io,registry.org=changeme,registry.authtoken=changeme
```

The auth token field follows the format of Docker's X-Registry-Auth header. For credential-based
logins such as Docker Hub and Quay, use

```
$ echo '{"username":"jdoe","password":"secret","email":"jdoe@acme.com"}' | base64
```

For token-based logins such as Google Container Registry and Amazon ECR, use

```
$ echo '{"registrytoken":"9cbaf023786cd7"}' | base64
```

If you're looking to build from source or get started hacking on Prow, please see the
[hacking guide][hacking] for more information.

### Use It!

Climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!

# License

This software is covered under the Apache v2.0 license. You can read the license [here][license].

This software contains a large amount of code from [Helm][], which is also covered by the Apache
v2.0 license.


[Getting Started]: docs/getting-started.md
[hacking]: docs/contributing/hacking.md
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[license]: LICENSE
