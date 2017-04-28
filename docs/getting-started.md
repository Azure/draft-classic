# Getting Started

This document shows how to deploy a "Hello World" app with Draft. To follow along, be sure you
have completed the [Hacking on Draft](contributing/hacking.md) guide.

## App setup

Let's create a sample Python app using [Flask](http://flask.pocoo.org/).

```shell
$ mkdir hello-world
$ cd hello-world
$ cat <<EOF > hello.py
from flask import Flask

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello World!\n"

if __name__ == "__main__":
    app.run()
EOF
$ echo "flask" > requirements.txt
$ ls
hello.py  requirements.txt
```

## Draft Create

We need some "scaffolding" to deploy our app into a [Kubernetes][] cluster. Draft can create a
[Helm][] chart, a `Dockerfile` and a `draft.toml` with `draft create`:

```shell
$ draft create
--> Default app detected
--> Ready to sail
```

## App-specific Modifications

The `chart/` and `Dockerfile` assets created by Draft default to a basic [nginx][]
configuration. For this exercise, let's replace the `Dockerfile` with one more Pythonic:

```shell
$ cat Dockerfile
FROM nginx:latest
$ cat <<EOF > Dockerfile
FROM python:onbuild

CMD [ "python", "./hello.py" ]

EXPOSE 80
EOF
```

This `Dockerfile` harnesses the [python:onbuild](https://hub.docker.com/_/python/) image, which
will install the dependencies in `requirements.txt` and copy the current directory
into `/usr/src/app`. And to align with the service values in `chart/values.yaml`, this Dockerfile
exposes port 80 from the container.

## Draft Up

Now we're ready to deploy `hello.py` to a Kubernetes cluster.

Draft handles these tasks with one `draft up` command:

- builds a Docker image from the `Dockerfile`
- pushes the image to a registry
- installs the Helm chart under `chart/`, referencing the Docker registry image


Let's use the `--watch` flag so we can let this run in the background while we make changes later on...

```shell
$ draft up --watch
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
  export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl port-forward $POD_NAME 8080:80

Watching local files for changes...
```

## Interact with Deployed App

Using the handy output that follows successful deployment, we can now contact our app:

```shell
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80
```

Oops! When we curl our app at `localhost:8080` we don't see "Hello World".  Indeed, if we were to check in on the application pod we would see its `Readiness` and `Liveness` checks failing:

```shell
$ curl localhost:8080
curl: (52) Empty reply from server
$ kubectl describe pod $POD_NAME
Name:		hello-world-hello-world-2214191811-gt5s7
...
Events:
  FirstSeen	LastSeen	Count	From			SubObjectPath			Type		Reason		Message
  ---------	--------	-----	----			-------------			--------	------		-------
  2m		2m		1	{default-scheduler }					Normal		Scheduled	Successfully assigned hello-world-hello-world-2214191811-gt5s7 to minikube
...
  1m		17s		5	{kubelet minikube}	spec.containers{hello-world}	Warning		Unhealthy	Liveness probe failed: Get http://172.17.0.9:80/: dial tcp 172.17.0.9:80: getsockopt: connection refused
  2m		7s		13	{kubelet minikube}	spec.containers{hello-world}	Warning		Unhealthy	Readiness probe failed: Get http://172.17.0.9:80/: dial tcp 172.17.0.9:80: getsockopt: connection refused
```

## Update App

Ah, of course.  We need to change the `app.run()` command in `hello.py` to explicitly run on port 80 so that our app handles connections where we've intended:

```shell
$ cat <<EOF > hello.py
from flask import Flask

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello World!\n"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)
EOF
```

## Draft Up(grade)

Now if we watch the terminal that we initially called `draft up --watch` with, Draft will notice that there were changes made locally and call `draft up` again. Draft then determines that the Helm release already exists and will perform a `helm upgrade` rather than attempting another `helm install`:

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
  export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl port-forward $POD_NAME 8080:80
```

## Great Success

Every `draft up` recreates the application pod, so we need to re-run the export and port-forward
steps from above.

```shell
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80
```

Now when we navigate to `localhost:8080` we see our app in action!  A beautiful `Hello World!` greets us.  Our first app has been deployed to our [Kubernetes][] cluster via Draft.

## Extra Credit

As a bonus section, we can utilize [Draft packs](packs.md) to create a Python-specific "pack" for scaffolding future Python apps.  As seen in the packs [doc](packs.md), as long as we place our custom pack in `$(draft home)/packs`, Draft will be able to find and use them.

For now, let's just copy over our `Dockerfile` and `chart/` assets to this location, to be built on at a later date:

```shell
$ mkdir -p $(draft home)/packs/python
$ cp -r Dockerfile chart $(draft home)/packs/python/
```

Now when we wish to create and deploy our new-fangled "Hello Universe" app, we can use our `python` Draft pack:

```shell
$ mkdir hello-universe
$ cp hello.py requirements.txt hello-universe
$ cd hello-universe
$ perl -i -0pe 's/World/Universe/' hello.py
$ draft create --pack=python
--> Ready to sail
$ draft up
...
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-universe-hello-universe" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80 &>/dev/null &
$ curl localhost:8080 && echo
Hello Universe!
```

[Helm]: https://github.com/kubernetes/helm
[nginx]: https://nginx.org/en/
[Kubernetes]: https://kubernetes.io/
