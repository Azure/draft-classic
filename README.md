![Draft Logo](./docs/img/draft-logo.png)

# Draft: Streamlined Kubernetes Development

[![Build Status](https://ci.deis.io/buildStatus/icon?job=Azure/draft/master)](https://ci.deis.io/job/Azure/job/draft/job/master/)

_NOTE: Draft is experimental and does not have a stable release yet._

Draft handles the heavy lifting involved in taking source code and deploying it to Kubernetes:

- Builds a container image from application source code
- Pushes the image to a registry
- Packages a [Helm][] chart from application source code
- Installs the chart to Kubernetes, deploying the application

## Installation

Review the [Installation Guide][Installation Guide] to configure and install Draft on to your Kubernetes cluster.

## Take Draft for a Spin

Climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!

## Contributing

If you're looking to build from source or get started hacking on Draft, please see the
[hacking guide][hacking] for more information.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## License

This software is covered under the MIT license. You can read the license [here][license].

This software contains code from Heroku Buildpacks, which are also covered by the MIT license.

This software contains code from [Helm][], which is covered by the Apache v2.0 license.

You can read third-party software licenses [here][Third-Party Licenses].


[Installation Guide]: docs/install.md
[Getting Started]: docs/getting-started.md
[hacking]: docs/contributing/hacking.md
[`helm` v2.4.2]: https://github.com/kubernetes/helm/releases/tag/v2.4.2
[Helm]: https://github.com/kubernetes/helm
[Installing Helm]: https://github.com/kubernetes/helm/blob/master/docs/install.md
[Kubernetes]: https://kubernetes.io/
[license]: LICENSE
[Third-Party Licenses]: NOTICE
