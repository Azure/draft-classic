When developing locally with minkube:

In one terminal window:
```console
$ helm init
$ tiller_deploy=$(kubectl get po -n kube-system -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep "tiller")
$ kubectl port-forward $tiller_deploy 44134:44134 -n kube-system
```

In another terminal window:
```console
$ eval $(docker-machine env default)
$ docker run -dp 5000:5000 registry
$ draftd start --listen-addr="127.0.0.1:44135" --registry-auth="e30K" --tiller-uri=":44134" --basedomain=k8s.local --local```
```

In another terminal you can run draft commands after the following step:
```
export DRAFT_HOST=127.0.0.1:44135
```

*** Special thanks to Brian Hardock for these steps ***
