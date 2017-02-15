# Prow: Streamlined Kubernetes Development

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

Climb aboard and explore the [lifecycle of an app deployed with Prow][Getting Started] - you'll soon be sailing!

[Getting Started]: docs/getting-started.md
[hacking]: docs/hacking.md
[Kubernetes]: https://kubernetes.io/
[Helm]: https://github.com/kubernetes/helm
