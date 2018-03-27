# Getting Started

This document shows how to deploy a "Hello World" app with Draft. If you haven't done so already, be sure you have Draft installed according to the [Installation Guide][Installation Guide].

## App setup

There are multiple example applications included within the [examples directory](../examples). For this walkthrough, we'll be using the [python example application](../examples/example-python) which uses [Flask](http://flask.pocoo.org/) to provide a very simple Hello World webserver.

```shell
$ cd examples/example-python
```

## Draft Create

We need some "scaffolding" to deploy our app into a [Kubernetes](https://kubernetes.io/) cluster. Draft can create a [Helm](https://helm.sh/) chart, a `Dockerfile` and a `draft.toml` with `draft create`:

```shell
$ draft create
--> Draft detected the primary language as Python with 97.267760% certainty.
--> Ready to sail
$ ls -a
.draftignore  Dockerfile  app.py  chart/  draft.toml  requirements.txt
```

The `chart/` and `Dockerfile` assets created by Draft default to a basic Python configuration. This `Dockerfile` harnesses the [python:onbuild](https://hub.docker.com/_/python/) image, which will install the dependencies in `requirements.txt` and copy the current directory into `/usr/src/app`. And to align with the service values in `chart/values.yaml`, this Dockerfile exposes port 80 from the container.

The `draft.toml` file contains basic configuration about the application like the name, the repository, which namespace it will be deployed to, and whether to deploy the app automatically when local files change.

```shell
$ cat draft.toml
[environments]
  [environments.development]
    name = "example-python"
    namespace = "default"
    wait = false
    watch = false
    watch-delay = 2
    override-ports = ["8080:8080", "9229:9229"]
    auto-connect = false    
```

See [DEP 6](reference/dep-006.md) for more information and available configuration on the `draft.toml`.

A `.draftignore` file is created as well for elements we want to exclude tracking on `draft up` when watching for changes. The syntax is identical to [helm's .helmignore file](https://github.com/kubernetes/helm/blob/master/pkg/repo/repotest/testdata/examplechart/.helmignore).

## Draft Up

Now we're ready to deploy this app to a Kubernetes cluster. Draft handles these tasks with one `draft up` command:

- reads configuration from `draft.toml`
- compresses the `chart/` directory and the application directory as two separate tarballs
- builds the image using `docker`
- `docker` pushes the image to the registry specified in `draft.toml` (or in `draft config get registry`, if set)
- `draft` instructs helm to install the chart, referencing the image just built

```shell
$ draft up
Draft Up Started: 'example-python'
example-python: Building Docker Image: SUCCESS ⚓  (73.0991s)
example-python: Pushing Docker Image: SUCCESS ⚓  (69.1425s)
example-python: Releasing Application: SUCCESS ⚓  (0.6875s)
example-python: Build ID: 01BSY5R8J45QHG9D3B17PAXMGN
```

## Interact with the Deployed App

Now that the application has been deployed, we can connect to our app.

```shell
$ draft connect
Connect to python:8080 on localhost:8080
172.17.0.1 - - [13/Sep/2017 19:10:09] "GET / HTTP/1.1" 200 -
```

`draft connect` is the command used to interact with the application deployed on your cluster. It works by creating proxy connections to the ports exposed by the containers in your pod, while also streaming the logs from all containers.

In another terminal window, we can connect to our app using the address displayed from `draft connect`'s output.

```shell
$ curl localhost:8080
Hello, World!
```

Once you're done playing with this app, cancel out of the `draft connect` session using CTRL+C.

> Note that you can use the flag `draft up --auto-connect` in order to have the application automatically connect once the deployment is done.

> You can customize the local ports for the `draft connect` command either through the `-p` flag or through the `override-ports` field in `draft.toml`. More info in [dep-007.md][dep007]

## Update the App

Now, let's change the output in `app.py` to output "Hello, Draft!" instead:

```shell
$ cat <<EOF > app.py
from flask import Flask

app = Flask(__name__)

@app.route("/")
def hello():
    return "Hello, Draft!\n"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)
EOF
```

## Draft Up(grade)

When we call `draft up` again, Draft determines that the Helm release already exists and will perform a `helm upgrade` rather than attempting another `helm install`:

```shell
$ draft up
Draft Up Started: 'example-python'
example-python: Building Docker Image: SUCCESS ⚓  (13.0127s)
example-python: Pushing Docker Image: SUCCESS ⚓  (16.0272s)
example-python: Releasing Application: SUCCESS ⚓  (0.5533s)
example-python: Build ID: 01BSYA4MW4BDNAPG6VVFWEPTH8
```

We should notice a significant faster build time here. This is because Docker is caching unchanged layers and only compiling layers that need to be re-built in the background.

## Great Success!

Now when we run `draft connect` and open the local URL using `curl` or our browser, we can see our app has been updated!

```shell
$ curl localhost:8080
Hello, Draft!
```

[Installation Guide]: install.md
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[Python]: https://www.python.org/
[dep007]: reference/dep-007.md