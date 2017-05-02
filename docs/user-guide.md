# Draft User Guide

This guide is an "advanced" setup on how a user can wire up their app's repository with draft.

## draft.toml

Draft configuration is stored in `draft.toml` in your app's root directory. [TOML][] is a minimal
configuration file format that is easy to read due to obvious semantics. The format of this file is
as follows:

```
[environments]
    [environments.development]
    name = "draft-dev"
    set = ["foo=bar", "car=star"]
    watch = true
    watch_delay = 1

    [environments.staging]
    name = "draft-dev"
    namespace = "kube-system"
    build_tar = "build.tar.gz"
    chart_tar = "chart.tar.gz"
```

Let's break it down by section:

```
[environments]
```

The root of the TOML file. Each definition under this node is considered an "environment". More on
that in a second.

```
    [environments.development]
```

This is the environment name. Applications deployed by Draft can be configured in different manners
based on the present environment. By default, `draft up` deploys using the `development` environment,
but this can be tweaked by either setting `$DRAFT_ENV` or by supplying the environment name at
runtime using `draft up --environment=staging`.

```
    name = "draft"
    namespace = "kube-system"
    build_tar = "build.tar.gz"
    chart_tar = "chart.tar.gz"
    set = ["foo=bar", "car=star"]
    wait = false
    watch = false
    watch_delay = 2
```

Here is a run-down on each of the fields:

 - `name`: name of the application. This will map directly with the name of the Helm release.
 - `namespace`: the kubernetes namespace where the application will be deployed.
 - `build_tar`: path to a gzipped build tarball. `chart_tar` must also be set.
 - `chart_tar`: path to a gzipped chart tarball. `build_tar` must also be set.
 - `set`: set custom Helm values.
 - `wait`: specifies whether or not to wait for all resources to be ready when Helm installs the chart.
 - `watch`: whether or not to deploy the app automatically when local files change.
 - `watch_delay`: the delay for local file changes to have stopped before deploying again (in seconds).

Note: All updates to the `draft.toml` will take effect the next time `draft up --environment=<affected environment>` is invoked _except_ the `namespace` key/value pair. Once a deployment has occurred in the original namespace, it won't be transferred over to another.


[toml]: https://github.com/toml-lang/toml
