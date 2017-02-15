# Prow: Streamlined Kubernetes Development

_NOTE: Prow is experimental and does not have a stable release yet._

Prow is a tool for developers to create and deploy cloud-native applications on [Kubernetes][]. It
can be used to

- Install any application onto Kubernetes; Prow is language-agnostic and customizable.
- Work with any Kubernetes cluster that supports [Helm][]
- Intelligently manage and upgrade Helm releases
- Create Helm Charts that can be packaged for delivery to your ops team

## Prow, the Elevator Pitch

Prow is a tool to install applications onto a Kubernetes cluster.

Prow has two parts: a client (prow) and a server (prowd). Prowd runs inside of your Kubernetes
cluster alongside [Tiller][helm] and packages a Docker image, uploads it to a Docker Registry and
installs a chart using that image. Prow runs on your laptop, CI/CD, or wherever you want it to run
and uploads your application to Prowd via `prow up`.

## Usage

Because Prow is currently experimental, there is no stable release out yet and users are expected
to build the project from source until we get some automation up in here. Please see
[this doc][hacking] to get started hacking on Prow.

Now that's all done, start from a source code repository and let Prow create the Kubernetes
packaging:

```
$ cd my-app
$ ls
app.py
$ prow create --pack=python
--> Created chart/
--> Created Dockerfile
--> Ready to sail
```

Now start it up!

```
$ prow up
--> Building Dockerfile
--> Pushing 127.0.0.1:5000/my-app:08db751fccc4fc01c8b2ee96111525ea43f064f2
--> Deploying to Kubernetes
    Release "my-app" does not exist. Installing it now.
--> code:DEPLOYED
```

That's it! You're now running your Python app in a [Kubernetes][] cluster.

Behind the scenes, Prow handles the heavy lifting:

- Builds a container image from application source code
- Pushes the image to a registry
- Packages a [Helm][] chart from application source code
- Installs the chart to Kubernetes, deploying the application

After deploying, you can run `prow up` again to create new releases when
application source code has changed.

## Next Steps

Looking for a more detailed, hands-on exercise exploring the lifecycle of an app deployed with Prow?  Check out the [Getting Started][] doc and you'll soon be sailing!

[Getting Started]: docs/getting-started.md
[hacking]: docs/hacking.md
[Kubernetes]: https://kubernetes.io/
[Helm]: https://github.com/kubernetes/helm
