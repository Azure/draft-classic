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
    build_tar = "build.tar.gz"
    chart_tar = "chart.tar.gz"
    namespace = "kube-system"
    set = ["foo=bar", "car=star"]
    wait = false
    watch = false
    watch_delay = 2
```

These fields all behave the exact same as they do as the option flags on `draft up`. See
`draft up --help` for more information.

Note:  All updates to the `draft.toml` will take effect the next time `draft up --environment=<affected environment>` is invoked _except_ the `namespace` key/value pair.  Once a deployment has occurred in the original namespace, it won't be transferred over to another.


[toml]: https://github.com/toml-lang/toml
