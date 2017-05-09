![Draft Logo](./docs/img/draft-logo.png)

# Draft: Streamlined Kubernetes Development

[![Build Status](https://ci.deis.io/buildStatus/icon?job=Azure/draft/master)](https://ci.deis.io/job/Azure/job/draft/job/master/)

_NOTE: Draft is experimental and does not have a stable release yet._

Draft handles the heavy lifting involved in taking source code and deploying it to Kubernetes:

- Builds a container image from application source code
- Pushes the image to a registry
- Packages a [Helm][] chart from application source code
- Installs the chart to Kubernetes, deploying the application

## Usage

### Install Draft

Because Draft is currently experimental, there is no stable release out yet and users are expected
to be using the latest build of Draft for testing. Canary releases of the Draft client can be found
at the following links:

 - [Linux amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)
 - [Windows amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-windows-amd64.tar.gz)

Unpack the Draft binary and add it to your PATH and you are good to go!

To install the server-side of Draft, use `draft init` with your credentials to let Draft communicate
with a Docker registry:

```
$ draft init --set registry.url=docker.io,registry.org=changeme,registry.authtoken=changeme
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

If you're looking to build from source or get started hacking on Draft, please see the
[hacking guide][hacking] for more information.

### Use It!

Climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!

## Contributing

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## License

This software is covered under the MIT license. You can read the license [here][license].

This software contains code from Heroku Buildpacks, which are also covered by the MIT license.

This software contains code from [Helm][], which is covered by the Apache v2.0 license.

You can read third-party software licenses [here][Third-Party Licenses].


[Getting Started]: docs/getting-started.md
[hacking]: docs/contributing/hacking.md
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[license]: LICENSE
[Third-Party Licenses]: NOTICE
