# Getting Started

This document illustrates the workflow of deploying a sample "Hello World" app via Prow.  If intending to follow along, be sure to have met the prerequisites in the [Hacking guide][hacking.md#prerequisites].

## App setup

Here we'll set up a sample Python "Hello World" app using [Flask](http://flask.pocoo.org/).

```
$ mkdir hello-world
$ cd hello-world
$ cat <<EOF > hello.py
from flask import Flask
app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello World!"

if __name__ == "__main__":
    app.run()
EOF
$ echo "flask" > requirements.txt
$ ls
hello.py  requirements.txt
```

## Prow Create

We're now ready to create the scaffolding needed to deploy our app into a [Kubernetes][] cluster.  Prow handles the creation of a default `Dockerfile` and [Helm][Helm] chart directory in `chart/` with `prow create`:

```
$ prow create
--> Created chart/
--> Created Dockerfile
--> Ready to sail
```

## App-specific Modifications

Both the Dockerfile and chart assets created by Prow default to basic [nginx][nginx] configurations. However, for this exercise, all we need to do is update the `Dockerfile` with Python-specific details:

```
$ cat Dockerfile
FROM nginx:latest
$ cat <<EOF > Dockerfile
FROM python:onbuild

CMD [ "python", "./hello.py" ]

EXPOSE 80
EOF
```

The resultant image build will harness the [python:onbuild](https://hub.docker.com/_/python/) image's handling of installing the dependencies specified by our `requirements.txt` and then copying the current directory into `/usr/src/app`.  In addition, to align with the corresponding service values in `chart/values.yaml` (see `service.externalPort` and `service.internalPort`), we've chosen port 80 to be exposed by the container.

## Prow Up

We're now ready to deploy our application into a Kubernetes cluster.  Prow will handle building the Docker image, pushing the image to the internal registry and then installing the chart described in `chart/`, having updated the appropriate values in `chart/values.yaml` to reference said internal registry image.

```
$ prow up
--> Building Dockerfile
Step 1 : FROM nginx:latest
latest: Pulling from library/nginx
5040bd298390: Already exists
333547110842: Pulling fs layer
4df1e44d2a7a: Pulling fs layer
4df1e44d2a7a: Verifying Checksum
4df1e44d2a7a: Download complete
333547110842: Verifying Checksum
333547110842: Download complete
333547110842: Pull complete
4df1e44d2a7a: Pull complete
Digest: sha256:f2d384a6ca8ada733df555be3edc427f2e5f285ebf468aae940843de8cf74645
Status: Downloaded newer image for nginx:latest
 ---> cc1b61406712
Successfully built cc1b61406712

--> Pushing 127.0.0.1:5000/hello-world:6c69b0e886fa89f49330c8f7900d02338ad47bc2
The push refers to a repository [127.0.0.1:5000/hello-world]

--> Deploying to Kubernetes
    Release "hello-world" does not exist. Installing it now.
--> code:DEPLOYED notes:"1. Get the application URL by running these commands:\n  export POD_NAME=$(kubectl get pods --namespace default -l \"app=hello-world-hello-world\" -o jsonpath=\"{.items[0].metadata.name}\")\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl port-forward $POD_NAME 8080:80\n"
```

_NOTE(vdice): update output above once https://github.com/deis/prow/issues/32 is fixed._

## Interact with Deployed App

Using the handy output that follows successful deployment, we can now contact our app:

```
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80
```

Oops! When we curl our app at `localhost:8080` we don't see "Hello World".  Indeed, if we were to check in on the application pod we would see its `Readiness` and `Liveness` checks failing:

```
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

```
$ cat <<EOF > hello.py
from flask import Flask
app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello World!"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)
EOF
```

## Prow Up(grade)

Now if we rerun `prow up` after making changes to our app, Prow will determine that the Helm release already exists and will hence perform a `helm upgrade` to this existing release rather than attempting another `helm install`:

```
$ prow up
--> Building Dockerfile
Step 1 : FROM python:onbuild
# Executing 3 build triggers...
Step 1 : COPY requirements.txt /usr/src/app/
 ---> Using cache
Step 1 : RUN pip install --no-cache-dir -r requirements.txt
 ---> Using cache
Step 1 : COPY . /usr/src/app
 ---> 721cc23a36b2
Step 2 : CMD python ./hello.py
 ---> Running in a11269445ce4
 ---> 4bae33ea2c9a
Step 3 : EXPOSE 80
 ---> Running in 0e8d3d99a840
 ---> 9c90b0445146
Successfully built 9c90b0445146

--> Pushing 127.0.0.1:5000/hello-world:f031eb675112e2c942369a10815850a0b8bf190e
The push refers to a repository [127.0.0.1:5000/hello-world]

--> Deploying to Kubernetes
--> code:DEPLOYED notes:"1. Get the application URL by running these commands:\n  export POD_NAME=$(kubectl get pods --namespace default -l \"app=hello-world-hello-world\" -o jsonpath=\"{.items[0].metadata.name}\")\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl port-forward $POD_NAME 8080:80\n"
```

## Great Success

As our every subsequent `prow up` invocation will recreate the application pod, we'll need to re-run the export and port-forward steps as done above.

```
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-world-hello-world" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80
```

Now when we navigate to `localhost:8080` we see our app in action!  A beautiful `Hello World!` greets us.  Our first app has been deployed to our [Kubernetes][] cluster via Prow.

## Extra Credit

As a bonus section, we can utilize [Prow packs](packs.md) to create a Python-specific "pack" for scaffolding future Python apps.  As seen in the packs [doc](packs.md), as long as we place our custom pack in `~/.prow/packs`, Prow will be able to find and use them.

For now, let's just copy over our `Dockerfile` and `chart/` assets to this location, to be built on at a later date:

```
$ mkdir -p ~/.prow/packs/python
$ cp -r Dockerfile chart ~/.prow/packs/python/
```

Now when we wish to create and deploy our new-fangled "Hello Universe" app, we can use our `python` Prow pack:

```
$ mkdir hello-universe
$ cp hello.py requirements.txt hello-universe
$ cd hello-universe
$ perl -i -0pe 's/World/Universe/' hello.py
$ prow create --pack=python
--> Created chart/
--> Created Dockerfile
--> Ready to sail
$ prow up
...
$ export POD_NAME=$(kubectl get pods --namespace default -l "app=hello-universe-hello-universe" -o jsonpath="{.items[0].metadata.name}")
$ kubectl port-forward $POD_NAME 8080:80 &>/dev/null &
$ curl localhost:8080 && echo
Hello Universe!
```

[Helm]: https://github.com/kubernetes/helm
[nginx]: https://nginx.org/en/
[Kubernetes]: https://kubernetes.io/
