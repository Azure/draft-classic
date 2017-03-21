# Prow User Guide

This guide is an "advanced" setup on how a user can wire up their app's repository with Prow.

## prow.yaml

Prow configuration is stored in `prow.yaml` in your app's root directory. The format of this file
is as follows:

```
environments:
  development:
    name: prow-dev
    set: ["foo=bar", "car=star"]
    watch: true
    watch_delay: 1
  staging:
  	name: prow-qa
    namespace: kube-system
    build-tar: build.tar.gz
    chart-tar: chart.tar.gz
    set: ["foo=bar", "car=star"]
    values: ["values/qa.yaml"]
    wait: true
```

Let's break it down by section:

```
environments:
```

The root of the YAML file. Each definition under this node is considered an "environment". More on
that in a second.

```
  development:
```

This is the environment name. Applications deployed by Prow can be configured in different manners
based on the present environment. By default, `prow up` deploys using the `development` environment,
but this can be tweaked by either setting `$PROW_ENV` or by supplying the environment name at
runtime using `prow up --environment=staging`.

```
    name: prow
    build_tar: build.tar.gz
    chart_tar: chart.tar.gz
    namespace: kube-system
    set: ["foo=bar", "car=star"]
    values: ["values/qa.yaml"]
    wait: false
    watch: false
    watch_delay: 2
```

These fields all behave the exact same as they do as the option flags on `prow up`. See
`prow up --help` for more information.
