# Install Guide

Get started with Draft in three easy steps:

1. Install CLI tools for Helm, Kubectl, [Minikube][] and Draft
2. Boot Minikube and install Draft
3. Deploy your first application

## Dependencies

In order to get started, you will need to fetch the following:

- the latest release of minikube
- the latest release of kubectl
- the latest release of Helm

All of the dependencies can be installed by the following:

```
$ brew cask install minikube
```

Afterwards, fetch [the latest release of Draft](https://github.com/Azure/draft/releases).

Installing Draft via Homebrew can be done using

```
$ brew tap azure/draft
$ brew install draft
```

Canary releases of the Draft client can be found at the following links:

 - [Linux amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-amd64.tar.gz)
 - [macOS amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)
 - [Windows amd64](https://azuredraft.blob.core.windows.net/draft/draft-canary-darwin-amd64.tar.gz)

It can also be installed with

```
$ brew install draft-canary
```

Alternative downloads:

- [Linux ARM](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-arm.tar.gz)
- [Linux x86](https://azuredraft.blob.core.windows.net/draft/draft-canary-linux-386.tar.gz)

Unpack the Draft binary and add it to your PATH.

## Boot Minikube

At this point, you can boot up minikube!

```
$ minikube start
Starting local Kubernetes v1.7.3 cluster...
Starting VM...
oving files into cluster...
Setting up certs...
Starting cluster components...
Connecting to cluster...
Setting up kubeconfig...
Kubectl is now configured to use the cluster.
```

Now that the cluster is up and ready, minikube automatically configures kubectl, the command line tool for Kubernetes, on your machine with the appropriate authentication and endpoint information.

```
$ kubectl cluster-info
Kubernetes master is running at https://192.168.99.100:8443

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

## Enable Minikube Add-ons

Now that we have minikube running, we can go ahead and enable the `registry` and `ingress`
add-ons.

The ingress add-on is used to allow inbound connections to reach the application.

The registry add-on is used to store the built docker container within the cluster.

You can enable the add-ons with

```console
$ minikube addons enable ingress
$ minikube addons enable registry
```

## Install Helm

Install Helm, a Kubernetes Package Manager, in your cluster. Helm manages the lifecycle of an application in Kubernetes, and it is also how Draft deploys an application to Kubernetes.

Installing Helm and setting it up is quite simple:

    $ helm init

Wait for Helm to come up and be in a `Ready` state. You can use `kubectl -n kube-system get deploy tiller-deploy --watch`
to wait for tiller to come up.

## Install Draft

Now that all the dependencies are set up, we can set up Draft by running this command:

    $ draft init

Follow through the prompts. Draft will read your local kube configuration and notice that it is
pointing at minikube. It will then install Draftd (the Draft server) communicating with the
installed registry add-on, ingress controller and Tiller (Helm server) instances.

## Configure Ingress Routes

Draft uses a wildcard domain to make accessing draft-created applications easier. To do so, it
specifies a custom host in the ingress from which tells the backing load balancer to route requests
based on the Host header.

When Draft was installed on Minikube, a base domain of `k8s.local` was used. To use this domain, you
can install `dnsmasq` to redirect all outgoing requests to `k8s.local` off to the Minikube cluster.

There are plenty of ways to install dnsmasq for MacOS users, but the easiest by far is to use
Homebrew.

    $ brew install dnsmasq

Once it's installed, you will want to point all outgoing requests to `k8s.local` to your minikube
instance.

```
$ echo 'address=/.k8s.local/`minikube ip`' > $(brew --prefix)/etc/dnsmasq.conf
$ sudo brew services start dnsmasq
```

This will start dnsmasq and make it resolve requests from `k8s.local` to your minikube instance's
IP address (usually some form of 192.168.99.10x), but now we need to point the operating system's
DNS resolver at dnsmasq to resolve addresses.

```
$ sudo mkdir /etc/resolver
$ echo nameserver 127.0.0.1 | sudo tee /etc/resolver/k8s.local
```

Afterwards, you will need to clear the DNS resolver cache so any new requests will go through
dnsmasq instead of hitting the cached results from your operating system.

```
$ sudo killall -HUP mDNSResponder
```

To verify that your operating system is now pointing all `k8s.local` requests at dnsmasq:

```
$ scutil --dns | grep k8s.local -B 1 -A 3
resolver #8
  domain   : k8s.local
  nameserver[0] : 127.0.0.1
  flags    : Request A records, Request AAAA records
  reach    : Reachable, Local Address, Directly Reachable Address
```

If you're on Linux, refer to [Arch Linux's fantastic wiki on dnsmasq][dnsmasq].

If you're on Windows, refer to [Acrylic's documentation][acrylic], which is another local DNS proxy
specifically for Windows. Just make sure that Acrylic is pointing at minikube through `k8s.local`.
You can use the above steps as a general guideline on how to set up Acrylic.

## Take Draft for a Spin

Once you've completed the above steps, you're ready to climb aboard and explore the
[Getting Started Guide][Getting Started] - you'll soon be sailing!


[acrylic]: http://mayakron.altervista.org/wikibase/show.php?id=AcrylicHome
[dnsmasq]: https://wiki.archlinux.org/index.php/dnsmasq
[Getting Started]: getting-started.md
[Ingress Guide]: ingress.md
[minikube]: https://github.com/kubernetes/minikube
