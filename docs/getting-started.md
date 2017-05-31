# Getting Started

This document shows how to deploy a "Hello World" app with Draft. If you havent done so already,
be sure you have Draft installed according to the [Installation Guide][Installation Guide].

## App setup

There are multiple example applications included within the [examples directory](../examples).
For this walkthrough, we'll be using the [python example application](../examples/python) which
uses [Flask](http://flask.pocoo.org/) to provide a very simple Hello World webserver.

```shell
$ cd examples/python
```

## Draft Create

We need some "scaffolding" to deploy our app into a [Kubernetes](https://kubernetes.io/) cluster. Draft can create a [Helm](https://helm.sh/) chart, a `Dockerfile` and a `draft.toml` with `draft create`:

```shell
$ draft create
--> Python app detected
--> Ready to sail
$ ls
Dockerfile  app.py  chart/  draft.toml  requirements.txt
```

The `chart/` and `Dockerfile` assets created by Draft default to a basic Python
configuration. This `Dockerfile` harnesses the [python:onbuild](https://hub.docker.com/_/python/)
image, which will install the dependencies in `requirements.txt` and copy the current directory
into `/usr/src/app`. And to align with the service values in `chart/values.yaml`, this Dockerfile
exposes port 80 from the container.

The `draft.toml` file contains basic configuration about the application like the name, which
namespace it will be deployed to, and whether to deploy the app automatically when local files
change.

```shell
$ cat draft.toml
[environments]
  [environments.development]
    name = "tufted-lamb"
    namespace = "default"
    watch = true
    watch_delay = 2
```

See [the Draft User Guide](user-guide.md) for more information and available configuration on the
`draft.toml`.

## Draft Up

Now we're ready to deploy `app.py` to a Kubernetes cluster.

Draft handles these tasks with one `draft up` command:

- reads configuration from `draft.toml`
- compresses the `chart/` directory and the application directory as two separate tarballs
- uploads the tarballs to `draftd`, the server-side component
- `draftd` then builds the docker image and pushes the image to a registry
- `draftd` instructs helm to install the Helm chart, referencing the Docker registry image just built

With the `watch` option set to `true`, we can let this run in the background while we make changes
later on...

```shell
$ draft up
--> Building Dockerfile
Step 1 : FROM python:onbuild
onbuild: Pulling from library/python
...
Successfully built 38f35b50162c
--> Pushing docker.io/microsoft/tufted-lamb:5a3c633ae76c9bdb81b55f5d4a783398bf00658e
The push refers to a repository [docker.io/microsoft/tufted-lamb]
...
5a3c633ae76c9bdb81b55f5d4a783398bf00658e: digest: sha256:9d9e9fdb8ee3139dd77a110fa2d2b87573c3ff5ec9c045db6009009d1c9ebf5b size: 16384
--> Deploying to Kubernetes
    Release "tufted-lamb" does not exist. Installing it now.
--> Status: DEPLOYED
--> Notes:
     
  http://tufted-lamb.example.com to access your application

Watching local files for changes...
```

## Interact with the Deployed App

Using the handy output that follows successful deployment, we can now contact our app.

```shell
$ curl http://tufted-lamb.example.com
Hello Draft!
```

When we `curl` our app, we see our app in action! A beautiful "Hello Draft!" greets us.  If not, make sure you've followed the [Ingress Guide](ingress.md).

## Update the App

Now, let's change the output in `app.py` to output "Hello Kubernetes!" instead:

```shell
$ cat <<EOF > app.py
from flask import Flask

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello Kubernetes!\n"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)
EOF
```

## Draft Up(grade)

Now if we watch the terminal that we initially called `draft up` with, Draft will notice that there
were changes made locally and call `draft up` again. Draft then determines that the Helm release
already exists and will perform a `helm upgrade` rather than attempting another `helm install`:

```shell
--> Building Dockerfile
Step 1 : FROM python:onbuild
...
Successfully built 9c90b0445146
--> Pushing docker.io/microsoft/tufted-lamb:f031eb675112e2c942369a10815850a0b8bf190e
The push refers to a repository [docker.io/microsoft/tufted-lamb]
...
--> Deploying to Kubernetes
--> Status: DEPLOYED
--> Notes:
     
  http://tufted-lamb.example.com to access your application
```

## Great Success!

Now when we run `curl http://tufted-lamb.example.com`, we can see our app has been updated and deployed to Kubernetes automatically!

[Installation Guide]: install.md
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[Python]: https://www.python.org/
