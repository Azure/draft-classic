# Prow: Streamlined Kubernetes Development

[![Build Status](http://drone.champagne.deis.com/api/badges/deis/prow/status.svg)](http://drone.champagne.deis.com/deis/prow)

_NOTE: Prow is experimental and does not have a stable release yet._

Prow handles the heavy lifting involved in taking source code and deploying it to Kubernetes:

- Builds a container image from application source code
- Pushes the image to a registry
- Packages a [Helm][] chart from application source code
- Installs the chart to Kubernetes, deploying the application

## Usage

### Build Prow

Because Prow is currently experimental, there is no stable release out yet and users are expected
to build the project from source until we get some automation up in here. Please see
[this doc][hacking] to get started hacking on Prow.

### Use It!

Climb aboard and explore the [Getting Started Guide][Getting Started] - you'll soon be sailing!

# License

This software is covered under the Apache v2.0 license. You can read the license [here][license].

This software contains a large amount of code from [Helm][], which is also covered by the Apache
v2.0 license.


[Getting Started]: docs/getting-started.md
[hacking]: docs/hacking.md
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[license]: LICENSE
