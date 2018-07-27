# Getting Started

This document shows how to deploy a "Hello World" application with Draft. If you haven't done so already, be sure you have Draft installed properly. This [Quickstart Guide](quickstart.md) is the perfect resource if you still need to install Draft.

## Application Setup

There are multiple example applications included within the [examples directory](../examples). For this walkthrough, we'll be using the [python example application](../examples/example-python) which uses [Flask](http://flask.pocoo.org/) to provide a very simple Hello World webserver.

```shell
$ cd examples/example-python
```

## Draft Create

We need some "scaffolding" to deploy our application into a [Kubernetes](https://kubernetes.io/) cluster. Draft can create a [Helm](https://helm.sh/) chart, a `Dockerfile`, and a `draft.toml` with `draft create`:

```shell
$ draft create
--> Draft detected Python (96.739130%)
--> Ready to sail
$ ls -a
.dockerignore     .draftignore      app.py            draft.toml
.draft-tasks.toml Dockerfile        charts/           requirements.txt
```

The `charts/` and `Dockerfile` assets created by Draft default to a basic Python configuration. This `Dockerfile` harnesses the [python:onbuild](https://hub.docker.com/_/python/) image, which will install the dependencies in `requirements.txt` and copy the current directory into `/usr/src/app`. To align with the `internalPort` service value in `charts/python/values.yaml`, this `Dockerfile` exposes port 8080 from the container.

The `draft.toml` file contains basic configuration details about the application like the name, the repository, which namespace it will be deployed to, and whether to deploy the application automatically when local files change.

```shell
$ cat draft.toml
[environments]
  [environments.development]
    name = "example-python"
    namespace = "default"
    wait = true
    watch = false
    watch-delay = 2
    auto-connect = false
    dockerfile = ""
    chart = ""
```

See [dep-006.md][dep006] for more information and available configuration on the `draft.toml` file.

A `.draftignore` file is created for elements we want to exclude tracking on `draft up` when watching for changes. The syntax is identical to [helm's .helmignore file](https://github.com/kubernetes/helm/blob/master/pkg/repo/repotest/testdata/examplechart/.helmignore).

```shell
$ cat .draftignore
*.swp
*.tmp
*.temp
.git*
```

A [`.dockerignore`](https://docs.docker.com/engine/reference/builder/#dockerignore-file) file is created to ensure the docker context ignores files and directories that are not necessary.

```shell
$ cat .dockerignore
Dockerfile
draft.toml
charts/
```

A `.draft-tasks.toml` file is also created. This file allows you to configure tasks to be run before `draft up` (`pre-up` tasks), after `draft up` (`post-up` tasks), or after `draft delete` (`cleanup` tasks). This file is empty by default. See [dep-008.md][dep008] for more information and available configuration on the `.draft-tasks.toml` file.

## Draft Up

Now we're ready to deploy this application to a Kubernetes cluster. Draft handles these tasks with one `draft up` command:

- reads configuration from `draft.toml`
- compresses the `charts/` directory and the application directory as two separate tarballs
- builds the image using `docker`
- instructs `docker` to push the image to the registry specified in `draft.toml` (or in `draft config get registry`, if set)
- instructs `helm` to install the chart, referencing the image just built

```shell
$ draft up
Draft Up Started: 'example-python': 01BSY5R8J45QHG9D3B17PAXMGN
example-python: Building Docker Image: SUCCESS ⚓  (52.1337s)
example-python: Releasing Application: SUCCESS ⚓  (0.5309s)
Inspect the logs with `draft logs 01BSY5R8J45QHG9D3B17PAXMGN`
```

> NOTE: You might see a `WARNING: no registry has been set` message if no container registry has been configured in draft. You can set a container registry using the `draft config set registry docker.io/myusername` command. If you'd prefer to silence this warning instead, you can run `draft config set disable-push-warning 1`. Users can also skip the push process entirely using the `--skip-image-push` flag.

To ensure your application deployed as expected, run `kubectl get pods` and take a look at the output.

```shell
$ kubectl get pods
NAME                                     READY     STATUS    RESTARTS   AGE
example-python-python-6755c4944d-zbgvj   1/1       Running   0          5s
```

> NOTE: If you're using Minikube and your `STATUS` shows an error such as `ErrImagePull` or `ImagePullBackOff`, make sure you've configured Draft to build images directly using Minikube's Docker daemon. You can do so by running `eval $(minikube docker-env)`. 

> INFO: For more information on installing and configuring Minikube for use with Draft, check out [the Minikube installation guide here](install-minikube.md).

## Interact with the Deployed Application

Now that the application has been deployed, we can connect to it using `draft connect`.

The `draft connect` command is used to interact with the application deployed on your cluster. It works by creating proxy connections to the ports exposed by the containers in your pod. It also streams the logs from all containers.

```shell
$ draft connect
Connect to python:8080 on localhost:54794
[python]:  * Environment: production
[python]:    WARNING: Do not use the development server in a production environment.
[python]:    Use a production WSGI server instead.
[python]:  * Debug mode: off
[python]:  * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit)
```

> NOTE: The `WARNING: Do not use the development server in a production environment` message is coming from Flask. The message is in regard to Flask's built-in web server and can safely be ignored for our test purposes here.

In this example, you can see that `draft connect` has proxied port 8080 from our container to port 54794 on localhost. We can now open a browser window or another terminal window and connect to our application using the address and port displayed from `draft connect`'s output.

```shell
$ curl localhost:54794
Hello, World!
```

> IMPORTANT: Your local port will likely be different than the one seen here.

> NOTE: If `localhost` does not resolve on your system, try `curl 127.0.0.1:<PORT>` instead.

Once you're done checking the application out, you can cancel out of the `draft connect` session using `CTRL+C`.

> NOTE: You can use the flag `draft up --auto-connect` in order to have the application automatically connect once the deployment is done.

> INFO: You can also customize the local ports for the `draft connect` command by using the `-p` flag or through the `override-ports` field in `draft.toml`. More information on this can be found in [dep-007.md][dep007].

## Update the Application

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
Draft Up Started: 'example-python': 01CEQ5H21BWSR5M8HTJ5BVXPYW
example-python: Building Docker Image: SUCCESS ⚓  (1.0010s)
example-python: Releasing Application: SUCCESS ⚓  (2.1236s)
Inspect the logs with `draft logs 01CEQ5H21BWSR5M8HTJ5BVXPYW`
```

We should notice a significantly faster build time here. This is because Docker is caching unchanged layers and only compiling layers that need to be rebuilt in the background.

## Great Success!

We can run `draft connect` again to set up a proxy to our application:

```shell
$ draft connect
Connect to python:8080 on localhost:54961
[python]:  * Environment: production
[python]:    WARNING: Do not use the development server in a production environment.
[python]:    Use a production WSGI server instead.
[python]:  * Debug mode: off
[python]:  * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit)
```

Once we have the address and port, we can connect again using `curl` in a new terminal window or by browsing to the host and port in a browser window:

```shell
$ curl localhost:54961
Hello, Draft!
```

We can see the application updated successfully!

## Draft Delete

If you're done testing this application, you can terminate and remove it from your Kubernetes cluster. To do so, run `draft delete`:

```shell
$ draft delete
app 'example-python' deleted
```

If you run `kubectl get pods` shortly after, you should see your application `STATUS` is `Terminating`:

```shell
$ kubectl get pods
NAME                                     READY     STATUS        RESTARTS   AGE
example-python-python-688fcf849f-8ddh7   1/1       Terminating   0          5m
```

Once the termination completes, a `kubectl get pods` will show that the application no longer exists in your Kubernetes cluster:

```shell
$ kubectl get pods
No resources found.
```

> IMPORTANT NOTE: The `draft delete` command should be run with **extreme care and caution** as it performs a termination and removal of the application from your Kubernetes cluster.

> INFO: The `draft delete` command does not any image(s) created for the deployment within your Docker registry.

[dep006]: reference/dep-006.md
[dep007]: reference/dep-007.md
[dep008]: reference/dep-008.md
