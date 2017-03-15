# Prow Design

This document outlines the general design principles of Prow; what it does, how it works, what it
can (and cannot) do, as well as define its intended audience and place in the market.

## What It Does

Prow is a tool for developing, organizing, packaging, deploying, and managing applications in the
Kubernetes ecosystem. It uses Helm for orchestration, but it provides a toolbox for developers
building applications on top of Kubernetes.

## How To Use It

Prow has two main commands:

- `prow create` takes your existing code and creates a new Kubernetes app
- `prow up` deploys a development copy of your app into a Kubernetes cluster.

## Start from a Dockerfile

Say you have an application that is already Dockerized, but not Helm-ified. Prow can start from
your existing Dockerfile and create a new app:

```
$ ls
Dockerfile
src/
$ prow create
--> Default app detected
--> Ready to sail
$ ls
chart/
Dockerfile
src/
```

In the example above, `prow create` constructed a new Helm Chart for you, and stored it alongside
your source code so that you can add it to version control, and even manually edit it.

## Start from Scratch

If you want to start with an empty Prow app, you can simply run `prow create` and it will scaffold
a chart and a Dockerfile for you.

## Start from Code

The example in the _Start from Scratch_ section of this document showed starting from existing
source code. In this case, you must use a [Prow _pack_](packs.md) (a preconfigured template for
your chart) to tell Prow how to use your source code. Prow provides some default packs, but it's
also easy to add custom ones.

## Start from an Existing Chart

If you've already created Kubernetes applications, you can start with an existing chart, and simply
begin using Prow. There are a few patterns you may need to follow to meet the expectations for
Prow, but this is a matter of a few minutes of work; not a complete refactoring.

In this case, you don't even need `prow create`. You can just create the directory structure and
sail onward.

## Running Your Code

When you run `prow up`, Prow deploys your code to your Kubernetes cluster for you. It does the
following:

- Packages your code using a `docker build`.
- Sends your code to a Docker Registry.
- Installs (or upgrades) your chart using Helm

And when you're done with development, Prow's "first class" objects are all supported by the
Kubernetes toolchain. Simply deploy the chart to your production cluster in whatever way suits you.

## Prow Packs

Prow creates new charts based on two pieces of data: The information it can learn from your
project's directory structure, and a pre-built "scaffold" called a _pack_.

A Prow pack contains one or two things:

- A Helm Chart scaffold
- A base Dockerfile

Prow ships with a `default` pack and a few basic alternative packs. The default pack simply builds
a default Helm Chart and a Dockerfile that points to `nginx:latest`.

From this, you can tailor packs to your specific needs.

- Build a language-specific or library-specific pack that, for example, packages a Ruby
 Rails app
- Build a pack tailored to your team's DevOps needs or designs
- Build a pack that leverages custom Kubernetes ThirdPartyResource
 types.

See [packs.md](packs.md) for more information.

## How Prow Works

This is a look behind the curtain. Here's how Prow works:

- Prow uses several existing components:
 - A Kubernetes cluster
 - The Helm Tiller server
 - A Docker Registry
 - A directory full of "packs" for specific templates
- `prow create` reads a scaffold out of the appropriate pack, creates the necessary file system
 objects, and writes some basic configuration into your chart.
- `prow up` delivers your chart into the Kubernetes cluster (think
 `helm upgrade --install my-app my-app`)

## Directory Structure

Imagine an app named `my-app`, which contains a Dockerfile and some source:

```
myapp/
 Dockerfile
 src/
   app.py
```

After running `prow create`, this directory would have a chart built for it:

```
myapp/
 chart/
   Chart.yaml
   templates/
     deployment.yaml
     service.yaml
   values.yaml
 Dockerfile
 src/
   app.py
```

The `chart/` directory is a complete Helm chart, tailored to your
needs.

Inside of the `values.yaml` file, Prow configures images for your chart:

```
image:
  registry: gcr.io
  org: bacongobbler
  name: myapp
  tag: 0.1.0
```

This information is then available to all of your Helm charts. (e.g. via `{{.Values.image.name}}`)

The contents of the `templates/` directory are determined by the particular Pack you've used.

### Questions and Answers

_Can I have multiple Prow charts in the same repo?_

At this time, no. You can however use a `requirements.yaml` in your chart to note what your chart
depends on.

_Can I modify the chart, or do I have to accept whatever the pack gives
me?_

You can modify the contents of the `chart/` directory as you wish.
Consider them part of your source code.

Keep in mind that there are three values injected from Prowd into the chart which you'll likely
want to use:

```
image:
  registry: quay.io          # the address of the registry
  org: bacongobbler          # the organization of the image
  name: myapp                # the name of the image
  tag: 08db751               # the release of the image in the registry
```

_How do I add an existing chart to Prow?_

Just copy (`helm fetch`) it into the `chart/` directory. You need to tweak the values file to
read from `image.registry`, `image.org`, `image.name` and `image.tag` if you want Prow to regenerate Docker
images for you. See above.

_How do I deploy applications to production?_

Prow is a developer tool. While you _could_ simply use `prow up` to do this, we'd recommend using
`helm package` in conjuction with a CI/CD pipeline.

Remember: You can always package a Prow-generated chart with `helm package chart/` and load the
results up to a chart repository, taking advantage of the existing Helm ecosystem.

## Other Architectural Considerations

Instead of a prowd HTTP server, we could spawn a prow pod "job" (via `prow up`) that runs only when
`prow up` is called. In that case, the `prow` client would be the main focal point for server-side
configuration. This has the advantage of requiring fewer resource demands server-side, but might
make the client implementation (and security story) significantly more difficult. Furthermore, it
might make two `prow up` operations between two clients differ (the "Works on My Machine!"
problem).

## User Personas and Stories

**Persona:** Inexperienced Kube Dev

This user wants to just work in their normal environment, but be able to deploy to Kubernetes
without any extra work.

**Persona:** Kubernetes App Developer

This user knows all about Kubernetes, but doesn't enjoy the hassle of scaffolding out the same old
stuff when testing changes. This user wants a _starting point_ to deploy in-progress changes to a
cluster.

- As a user, I want to create a new Kubernetes app...
 - from scratch
 - from existing code
 - from a Dockerfile
 - from a chart
 - from some Kubernetes manifest files
- As a user, I want my app deployed quickly to a dev cluster
- As a user, I want to code, and have the deployed version auto-update
- As a user, I want to be as ignorant of the internals as possible, but still be able to GSD.
- As a user, I want to be able to hand off code artifacts without having to prepare them.
