# Getting Started

This document shows how to deploy a "Hello World" app with Draft. To follow along, be sure you
have Draft up and installed according to the [README](../README.md#install-draft).

## App setup

There are multiple example applications included within the [examples directory](../examples).
For this walkthrough, we'll be using the [python example application](../examples/python) which
uses [Flask](http://flask.pocoo.org/) to provide a very simple Hello World webserver.

```shell
$ cd examples/python
```

## Draft Create

We need some "scaffolding" to deploy our app into a [Kubernetes][] cluster. Draft can create a
[Helm][] chart, a `Dockerfile` and a `draft.toml` with `draft create`:

```shell
$ draft create
--> Python app detected
--> Ready to sail
$ ls
chart  Dockerfile  draft.toml  hello.py  requirements.txt
```

The `chart/` and `Dockerfile` assets created by Draft default to a basic [Python][]
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
    name = "snug-lamb"
    watch = true
    watch_delay = 2
```

See [the Draft User Guide](user-guide.md) for more information and available configuration on the
`draft.toml`.

## Draft Up

Now we're ready to deploy `hello.py` to a Kubernetes cluster.

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
--> Pushing quay.io/deis/hello-world:5a3c633ae76c9bdb81b55f5d4a783398bf00658e
The push refers to a repository [quay.io/deis/hello-world]
...
5a3c633ae76c9bdb81b55f5d4a783398bf00658e: digest: sha256:9d9e9fdb8ee3139dd77a110fa2d2b87573c3ff5ec9c045db6009009d1c9ebf5b size: 16384
--> Deploying to Kubernetes
    Release "hello-world" does not exist. Installing it now.
--> Status: DEPLOYED
--> Notes:
     1. Get the application URL by running these commands:
     NOTE: It may take a few minutes for the LoadBalancer IP to be available.
           You can watch the status of by running 'kubectl get svc -w hello-world-hello-world'
  export SERVICE_IP=$(kubectl get svc --namespace default hello-world-hello-world -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo http://$SERVICE_IP:80

Watching local files for changes...
```

## Interact with the Deployed App

Using the handy output that follows successful deployment, we can now contact our app. Note that it
may take a few minutes before the load balancer is provisioned by Kubernetes. Be patient!

```shell
$ export SERVICE_IP=$(kubectl get svc --namespace default hello-world-hello-world -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
$ curl http://$SERVICE_IP
```

When we `curl` our app, we see our app in action! A beautiful "Hello World!" greets us.

## Update the App

Now, let's change the "Hello World!" output in `hello.py` to output "Hello Draft!" instead:

```shell
$ cat <<EOF > hello.py
from flask import Flask

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello Draft!\n"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)
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
--> Pushing quay.io/deis/hello-world:f031eb675112e2c942369a10815850a0b8bf190e
The push refers to a repository [quay.io/deis/hello-world]
...
--> Deploying to Kubernetes
--> Status: DEPLOYED
--> Notes:
     1. Get the application URL by running these commands:
     NOTE: It may take a few minutes for the LoadBalancer IP to be available.
           You can watch the status of by running 'kubectl get svc -w hello-world-hello-world'
  export SERVICE_IP=$(kubectl get svc --namespace default hello-world-hello-world -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo http://$SERVICE_IP:80
```

## Great Success!

Now when we run `curl http://$SERVICE_IP`, our first app has been deployed and updated to our
[Kubernetes][] cluster via Draft!


[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[Python]: https://www.python.org/
