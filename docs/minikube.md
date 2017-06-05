## Minikube

Draft requires a wildcard domain and an ingress controller. Here how to set everything
up when using [Minikube][Minikube]

### Setup the certificate

Create a self-signed certificate using the following command. Don't forget to change
the `*.change.me` value to your own wildcard domain that you have used to start Draft
as mentionned in the [Installation Guide][Installation Guide].

```
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./tls.key -out ./tls.crt -subj "/CN=*.change.me
```

This will create two files, `tls.key` and `tls.crt`, that you will need to copy to the chart directory of
your project.

Next, create a secret config file to contain the certificate and key. In your `chart/templates`
directory create the file `cert-secret.yaml` with the following content

```
apiVersion: v1
kind: Secret
metadata:
  name: {{ template "fullname" . }}-tls
  labels:
    chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
type: Opaque
data:
  tls.crt: |-
    {{ .Files.Get "tls.crt" | b64enc }}
  tls.key: |-
    {{ .Files.Get "tls.key" | b64enc }}
```

### Ingress controller

Make sure you have enabled the ingress addon on your minikube instance

```
$ minikube addons enable ingress`
```

In your `chart/templates` directory, create or edit the `ingress.yaml` file to include
the tls section in the spec. Note that the list of hosts must contain the same
value as the host in the rules.

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: {{ template "fullname" . }}
  labels:
    chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
spec:
  tls:
  - hosts:
    - {{ .Release.Name }}.{{ .Values.basedomain }}
    secretName: {{ template "fullname" . }}-tls
  rules:
  - host: {{ .Release.Name }}.{{ .Values.basedomain }}
    http:
      paths:
      - path: /
        backend:
          serviceName: {{ template "fullname" . }}
          servicePort: {{ .Values.service.externalPort }}
```

You can now use Draft, on Minikube, as documented.


### Local experience

When using minikube, we can have a fully local experience without having to
refer to an external docker registry.

If you wish to start working with draft without having an external registry
setup, you can use the following procedure:


When you perform the `up` command, use the `--skip-push-image` option.
This will instruct the server to skip the image push and use only the local
image in the daemon.

You must make sure that the image pull policy, in your `deployment.yaml` file,
is set to `IfNotPresent` and not `Always`. If not, your must change or override
it through a custom values.yaml.

```
$ draft up --skip-image-push -f minikube-values.yaml
```


You can start draft with the default credentials but you must specify your
domain

```
draft init --set basedomain=change.me
```

[Installation Guide]: install.md
[Installation Guide]: install.md
[Minikube][https://github.com/kubernetes/minikube]

