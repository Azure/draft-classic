# Prow: Cloud Native Development

Prow is a tool for building Kubernetes-aware applications with ease. You bring the source code and Prow will negotiate the Kubernetes waters for you.


**ROUGH DRAFT**

## TL;DR

Start from your source code repository and let Prow transform it for
Kubernetes:

```
$ cd my-app
$ ls
app.py
$ prow new my-app --pack=python
--> Created ./charts/my-app
--> Created ./charts/my-app/Dockerfile
--> Ready to sail
```

Now start it up!

```
$ prow up
--> Building Dockerfile
--> Pushing my-app:latest
--> Deploying to Kubernetes
--> Ready at 10.21.77.7:8080
```

That's it! You're now running your Python app in a Kubernetes cluster.

## What It Does

Prow is a tool for developing, organizing, packaging, deploying, and
managing applications in the Kubernetes ecosystem. It uses Helm for
orchestration. But it provides a toolbox for developers building
applications composed of one or more microservices.

## How To Use It

Prow has two main commands:

- `prow create` takes your existing code and creates a new Kubernetes app
- `prow up` deploys a development copy of your app into a Kubernetes
  cluster. Optionally, you can have it watch your code for changes,
  re-deploying automatically with each change.

## Start from a Dockerfile

Say you have an application that is already Dockerized, but not Helm-ified. Prow can start
from your existing Dockerfile and create a new app:

```
$ ls
Dockerfile
src/
$ prow create
--> Created ./chart
--> Ready to sail
$ ls
chart/
Dockerfile
src/
```

In the example above, `prow create` created a new application and
based configuration off of the Dockerfile. Prow constructed a new
Helm Chart for you, and stored it alongside your source code so
that you can add it to version control, and even manually edit it.

## Start from Scratch

If you want to start with an empty Prow app, you can simply run `prow
create` and it will scaffold a chart and a Dockerfile for you.

## Start from Code

The example in the _tl;dr_ section of this document showed starting from
existing source code. In this case, you must use a Prow _pack_ (a
preconfigured template for your chart) to tell Prow how to use your
source code. Prow provides some default packs, but it's also easy to add
custom ones.

## Start from an Existing Chart

If you've already created Kubernetes applications, you can start with an
existing chart, and simply begin using Prow. There are a few patterns
you may need to follow to meet the expectations for Prow, but this is a
matter of a few minutes of work; not a complete refactoring.

In this case, you don't even need `prow create`. You can just create the
directory structure and sail onward.

## Running Your Code

When you run `prow up`, Prow deploys your code to your Kubernetes
cluster for you. It does the following:

- Package your code using a Docker build.
- Send your code to the in-cluster Docker registry.
- Install (or upgrade) your chart using Helm
- (Optionally) watch your source code for changes, redeploying when
 necessary.

And when You're done with development, Prow's "first class" objects
are all supported by the Kubernetes toolchain. Simply deploy the chart
to your production cluster in whatever way suits you.

## Prow Packs

Prow creates new charts based on two pieces of data: The information it
can learn from your project's directory structure, and a pre-build
"scaffold" called a _pack_.

A Prow pack contains one or two things:

- A Helm Chart scaffold
- A base Dockerfile (optional; used if a project does not contain a Dockerfile)

Prow ships with a `default` pack and a few basic alternative packs. The
default pack simply builds a chart with a service and a replica set.

From this, you can tailor packs to your specific needs.

- Build a language-specific pack that, for example, packages a Ruby
 Rails app.
- Build a pack tailored to your team's DevOps needs or designs
- Build a pack that leverages custom Kubernetes ThirdPartyResource
 types.

## How Prow Works

This is a look behind the curtain. Here's how Prow works:

- Prow uses several existing components:
 - A Kubernetes cluster
 - The Helm Tiller server
 - An in-cluster Docker registry
 - A repository of "packs" for specific templates
- While it's up to you to find the Kubernetes cluster and install Tiller,
 Prow can install the rest for you should you so desire with `prow init`.
- `prow create` reads a scaffold out of the appropriate Pack, creates the
 necessary file system objects, and writes some basic configuration
 into your chart.
- `prow up` delivers your chart into the Kubernetes cluster (think `helm
 upgrade --install my-app my-app`)
- `prow up` can be configured to either start your app and leave it
 alone, or watch your local code for changes.

## Directory Structure

Imagine an app named `my-app`, which contains a Dockerfile and some
source:


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
name: myapp
version: 0.1.0
registry: gcr.io
```

This information is then available to all of your Helm charts. (e.g. via
`{{.Values.name}}`)

The contents of the `templates/` directory are determined by the
particular Pack you've used. The `default` pack just creates a
replicated service for you.

### Questions and Answers

_Can I have multiple Prow charts in the same repo?_

At this time, no.

_Can I modify the chart, or do I have to accept whatever the pack gives
me?_

You can modify the contents of the `chart/` directory as you wish.
Consider them part of your source code.

_How do I add an existing chart to Prow?_

Just copy (`helm fetch`) it into the `chart/` directory. You need to
tweak the values file if you want Prow to regenerate Docker images for
you.

_How do I deploy applications to production?_

Prow is a developer tool. While you _could_ simply use `prow up` to do
this, we'd recommend using `helm` in conjuction with a CI/CD pipeline or
whatever tooling you're most comfortable with.

Remember: You can package a Prow-generated chart with `helm package chart/`
and load the results up to a chart repository, taking advantage of the
existing Helm ecosystem.

## Other Architectural Considerations

- instead of a prowd HTTP server, we could spawn a prow pod "job" (via `prow up`)
 that runs only when `prow up` is called. In that case, the `prow` client
 would be the main focal point for server-side configuration. This has the
 advantage of requiring fewer resource demands server-side, but might make the
 client implementation (and security story) significantly more difficult. Furthermore,
 it might make two `prow up` operations between two clients differ
 (the "Works on My Machine!" problem).

## User Personas and Stories

**Persona:** Inexperienced Kube Dev

This user wants to just work in their normal environment, but be able to
deploy to Kubernetes without any extra work.

**Persona:** Kubernetes App Developer

This user knows all about Kubernetes, but doesn't enjoy the hassle of
scaffolding out the same old stuff. This user wants a _starting point_.

- As a user, I want to create a new Kubernetes app...
 - from scratch
 - from existing code
 - from a Dockerfile
 - from a chart
 - from some Kubernetes manifest files
- As a user, I want my app deployed quickly to a dev cluster
- As a user, I want to code, and have the deployed version auto-update
- As a user, I want to be as ignorant of the internals as possible, but
 still be able to GSD.
- As a user, I want to be able to hand off code artifacts without having
 to prepare them.

