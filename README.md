# Prow: Streamlined Kubernetes Development

_NOTE: Prow is experimental and does not have a stable release yet._

Prow is a tool for developers to create cloud-native applications on [Kubernetes][].

## Usage

_NOTE: These instructions are temporary until there is a CLI._

Here's an example of installing Prow and creating an application:

```
$ go version
go version go1.7 linux/amd64
$ # install prowd
$ helm init
$ export IMAGE_PREFIX=bacongobbler
$ make info
Build tag:       git-abc1234
Registry:        quay.io
Immutable tag:   quay.io/deis/prowd:git-abc1234
Mutable tag:     quay.io/deis/prowd:canary
$ make docker-build docker-push
$ cat chart/values.yaml | grep repository
  repository: quay.io/deis/prowd
$ helm install ./chart --namespace prow
$ cd tests/testdata/example-dockerfile-http
$ prow up
--> Building Dockerfile
--> Pushing 127.0.0.1:5000/example-dockerfile-http:fc8c34ba4349ce3771e728b15ead2bb4c81cb9fd
--> Deploying to Kubernetes
    Release "example-dockerfile-http" does not exist. Installing it now.
--> code:DEPLOYED
$ helm list
NAME                      REVISION   UPDATED                     STATUS      CHART
example-dockerfile-http   1          Tue Jan 24 15:04:27 2017    DEPLOYED    example-dockerfile-http-1.0.0
$ kubectl get po
NAME                                       READY     STATUS             RESTARTS   AGE
example-dockerfile-http-3666132817-m2pkr   1/1       Running            0          30s
```

To confirm that the image deployed to Kubernetes matches the local archive:

```
$ shasum build.tar.gz | awk '{print $1}'
fc8c34ba4349ce3771e728b15ead2bb4c81cb9fd
$ kubectl get po example-dockerfile-http-3666132817-m2pkr -o=jsonpath='{.spec.containers[0].image}' | rev | cut -d ':' -f 1 | rev
fc8c34ba4349ce3771e728b15ead2bb4c81cb9fd
```

_NOTE: CLI usage will look like this when it is completed._

Start from a source code repository and let Prow create the Kubernetes packaging:

```
$ cd my-app
$ ls
app.py
$ prow create --pack=python
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
--> code:DEPLOYED
```

That's it! You're now running your Python app in a [Kubernetes][] cluster.

Behind the scenes, Prow handles the heavy lifting:

- Builds a container image from application source code
- Pushes the image to a registry
- Creates a [Helm][] chart for the image
- Installs the chart to Kubernetes, deploying the application

After deploying, you can run `prow up` again to create new releases when
application source code has changed. Or you can let Prow continuously rebuild
your application.

## Features

- Prow is language-agnostic
- Prow works with any Kubernetes cluster that supports [Helm][]
- Charts that Prow creates can be edited to suit your needs
- Charts can be packaged for delivery to your ops team

[Kubernetes]: https://kubernetes.io/
[Helm]: https://github.com/kubernetes/helm
