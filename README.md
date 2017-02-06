# Prow: Streamlined Kubernetes Development

_This project is experimental, and has not hit a stable release point_

Prow is a developer tool for creating cloud native applications for Kubernetes.

## Usage

_NOTE(bacongobbler): this is usage instructions to test while there's no client yet_

For now, this is the easiest way to test/run this locally on macOS:

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
NAME                                REVISION        UPDATED                         STATUS          CHART
example-dockerfile-http             1               Tue Jan 24 15:04:27 2017        DEPLOYED        example-dockerfile-http-1.0.0
$ kubectl get po
NAME                                       READY     STATUS             RESTARTS   AGE
example-dockerfile-http-3666132817-m2pkr   1/1       Running            0          30s
```

You can also confirm that the image deployed on kubernetes is the same as what was uploaded locally:

```
$ shasum build.tar.gz | awk '{print $1}'
fc8c34ba4349ce3771e728b15ead2bb4c81cb9fd
$ kubectl get po example-dockerfile-http-3666132817-m2pkr -o=jsonpath='{.spec.containers[0].image}' | rev | cut -d ':' -f 1 | rev
fc8c34ba4349ce3771e728b15ead2bb4c81cb9fd
```

_NOTE(bacongobbler): This is what the final CLI usage should look like_

Start from your source code repository and let Prow transform it for
Kubernetes:

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

That's it! You're now running your Python app in a Kubernetes cluster.

Behind the scenes, Prow is handling the heavy lifting for you:

- It builds a container image, and pushes it to a registry
- It creates a Helm chart for deploying into Kubernetes
- Using Helm, it deploys your release

From there, you can either let Prow continually rebuild your app, or you can
manually re-run Prow to update the existing app.

## Features

- Prow is language agnostic
- Prow will work with any Kubernetes cluster that supports Helm
- Once it scaffolds your chart, you can leave the chart alone, or you can edit
  it to your specific needs
- Charts can be packaged and delivered to your ops team

