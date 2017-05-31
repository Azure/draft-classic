## What is Ingress

Ingress is a way to route traffic from the internet to services within your Kubernetes cluster,
without creating a load-balancer for each service. For more information, review the
[Kubernetes Ingress Documentation][Kubernetes Ingress Documentation]

## Installing an Ingress Controller

While there are many ingress controllers available within the Kubernetes Community, for
simplicity, this guide will use the nginx-ingress from the stable helm charts, but you are
welcome to use any ingress controller.

These documents assume you are connected to a Kubernetes cluster running in a cloud-provider.

**NOTE:** If you are running in minikube, these steps will not work as desired. Additional documentation
about running draft on a minikube cluster will be made available shortly.

```shell
$ helm install stable/nginx-ingress --namespace=kube-system --name=nginx-ingress
```

After you've installed the nginx-ingress controller, wait for a Load Balancer to be created with:

```shell
$ kubectl --namespace kube-system get services -w nginx-ingress-nginx-ingress-controller
```

## Point a wildcard domain

Draft uses a wildcard domain to make accessing draft-created applications easier.

Using a domain that you manage, create a DNS wildcard `A Record` pointing to the ingress IP address.

**NOTE:** you are welcome to use `*.draft.example.com` or any other wildcard domain.

Remember the domain you use, it will be needed in the next step of installation as the `basedomain` passed to `draft init`.

| Name          | Type | Data                      |
|---------------|------|---------------------------|
| *.example.com | A    | `<ip address from above>` |

## Next steps

Once you have an ingress controller installed and configured on your cluster, you're ready
to install Draft. 

Continue with the [Installation Guide][Installation Guide]!


[Installation Guide]: install.md#install-draft
[Kubernetes Ingress Documentation]: https://kubernetes.io/docs/concepts/services-networking/ingress/
