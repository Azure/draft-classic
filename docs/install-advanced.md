# Advanced Install Guide

In certain situations, users are given a Kubernetes cluster with tighter security constraints such as

- only being able to deploy applications to a certain namespace
- running applications in an RBAC-enabled cluster

This document explains some of these situations as well as how it can be handled using draft.

## Running Tiller with RBAC enabled

To install Tiller in a cluster with RBAC enabled, a few additional steps are required to grant tiller access to deploy in namespaces.

[Helm's documentation](https://docs.helm.sh/using_helm/#role-based-access-control) on setting up Tiller in RBAC-enabled clusters is the best document on the subject, but for new users who just want to get started, let's create a new service account for tiller with the right roles to deploy to the `default` namespace.

```shell
$ kubectl create serviceaccount tiller --namespace kube-system
serviceaccount "tiller" created
```

Let's define a role that allows tiller to manage all resources in the `default` namespace. In `role-tiller.yaml`:

```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tiller-manager
  namespace: default
rules:
- apiGroups: ["", "extensions", "apps"]
  resources: ["*"]
  verbs: ["*"]
```

```shell
$ kubectl create -f role-tiller.yaml
role "tiller-manager" created
```

Now let's bind the service account to that role. In `rolebinding-tiller.yaml`:

```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tiller-binding
  namespace: default
subjects:
- kind: ServiceAccount
  name: tiller
  namespace: kube-system
roleRef:
  kind: Role
  name: tiller-manager
  apiGroup: rbac.authorization.k8s.io
```

```shell
$ kubectl create -f rolebinding-tiller.yaml
rolebinding "tiller-binding" created
```

We'll also need to grant tiller access to read configmaps in kube-system so it can store release information. In `role-tiller-kube-system.yaml`:

```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: kube-system
  name: tiller-manager
rules:
- apiGroups: ["", "extensions", "apps"]
  resources: ["configmaps"]
  verbs: ["*"]
```

```shell
$ kubectl create -f role-tiller-kube-system.yaml
role "tiller-manager" created
```

And the respective role binding. In `rolebinding-tiller-kube-system.yaml`:

```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tiller-binding
  namespace: kube-system
subjects:
- kind: ServiceAccount
  name: tiller
  namespace: kube-system
roleRef:
  kind: Role
  name: tiller-manager
  apiGroup: rbac.authorization.k8s.io
```

```shell
$ kubectl create -f rolebinding-tiller-kube-system.yaml
rolebinding "tiller-binding" created
```

Then, install tiller.

```shell
$ helm init --service-account=tiller
```

## Running Tiller in a namespace other than kube-system

To install Tiller in a namespace other than kube-system, a user can run `helm init` with the `--tiller-namespace` feature flag. This feature flag will deploy Tiller in that namespace.

There are a few prerequisites to make this work, however:

For minikube users, the setup documentation starts to look like this when deploying Tiller to the `draft` namespace:

```shell
$ minikube start
$ kubectl create namespace draft
$ helm init --tiller-namespace draft
```

From then on, you can interact with Tiller either with the `--tiller-namespace` flag or by setting `$TILLER_NAMESPACE` in your client's environment.

```
$ draft up --tiller-namespace draft
$ export TILLER_NAMESPACE=draft
$ draft up
```
