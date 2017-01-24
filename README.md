# Prow: Streamlined Kubernetes Development

_This project is experimental, and has not hit a stable release point_

Prow is a developer tool for creating cloud native applications for Kubernetes.

## Usage

_NOTE(bacongobbler): this is usage instructions to test while there's no client yet_

For now, this is the easiest way to test/run this locally on macOS:

```
$ # install prowd
$ helm init
$ export IMAGE_PREFIX=bacongobbler
$ make info
Build tag:       git-abc1234
Registry:        quay.io
Immutable tag:   quay.io/bacongobbler/prowd:git-abc1234
Mutable tag:     quay.io/bacongobbler/prowd:canary
$ make docker-build docker-push
$ helm install ./charts/prowd
$ # prepare the build context and chart tarballs
$ cd tests/testdata/example-dockerfile-http
$ tar czf build.tar.gz Dockerfile rootfs/
$ pushd charts/
$ tar czf ../charts.tar.gz example-dockerfile-http/
$ popd
$ # push the tarballs to prowd!
$ curl -XPOST -F release-tar=@build.tar.gz -F chart-tar=@charts.tar.gz http://k8s.local:44135/apps/foo
--> Building Dockerfile
--> Pushing 127.0.0.1:5000/foo:latest
--> Deploying to Kubernetes
    Release "foo" does not exist. Installing it now.
--> code:DEPLOYED
$ helm list
NAME            REVISION        UPDATED                         STATUS          CHART
foo             1               Tue Jan 24 15:04:27 2017        DEPLOYED        example-dockerfile-http-1.0.0
$ kubectl get po
NAME                   READY     STATUS             RESTARTS   AGE
foo-3666132817-m2pkr   1/1       Running            0          30s
```

_NOTE(bacongobbler): This is what the final CLI usage should look like_

Start from your source code repository and let Prow transform it for
Kubernetes:

```
$ cd my-app
$ ls
app.py
$ prow create my-app --pack=python
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

Behind the scenes, Prow is handling the heavy lifting for you:

- It uses your existing Dockerfile, or creates one for you
- It builds a container image, and pushes it to a registry
- It creates a Helm chart for deploying into Kubernetes
- Using Helm, it deploys your release

From there, you can either let Prow continually rebuild your app, or you can
manually re-run Prow to update the existing app.

## Features

- Prow is language agnostic
- Prow will work with any Kubernetes cluster that supports Helm
- One it scaffolds your chart, you can leave the chart alone, or you can edit
  it to your specific needs.
- Charts can be packaged and delivered to your ops team

