# Prow Starter Packs

`prow create` has an option flag called `--pack` that allows a user to bootstrap an application
from a defined starter pack in their local filesystem. Starter packs consist of a Dockerfile and
a chart that demonstrates best practices for an application of a given language pack. For example:

```
$ prow create --pack=rails
--> Created chart/
--> Created Dockerfile
--> Ready to sail
```

This document explains the starter pack format and provides basic guidance on creating your own
starter packs.

## The Starter Pack Structure

A starter pack is organized inside a directory in `~/.prow/packs`. Inside the pack's directory,
there will be a template `chart/` and a `Dockerfile` that will be injected into the application
when the starter pack is requested.

Inside this directory, Prow will expect a structure like this:

```
python/               # the name of the directory is the name of the starter pack
  chart/
    Chart.yaml        # A YAML file containing information about the chart
    LICENSE           # OPTIONAL: A plain text file containing the license for the chart
    README.md         # OPTIONAL: A human-readable README file
    values.yaml       # The default configuration values for this chart
    charts/           # OPTIONAL: A directory containing any charts upon which this chart depends.
    templates/        # OPTIONAL: A directory of templates that, when combined with values,
                      # will generate valid Kubernetes manifest files.
  detect              # OPTIONAL: An executable file for pack detection
  Dockerfile          # A Dockerfile for building the application
```

We could then run `prow create` with this pack like so:

```
$ prow create --pack=python
--> Created chart/
--> Created Dockerfile
--> Ready to sail
```

The easiest way to create and work with a starter pack is with the following commands:

```
$ cd ~/.prow/packs
$ mkdir python
$ cd python
$ helm create chart
Creating chart
$ echo "FROM python:onbuild" > Dockerfile
```

See [Helm's documentation on Charts][charts] for more information on the Chart file structure.

## Pack Detection

When `prow create` is executed on an application, prow starts iterating through the packs available
in `~/.prow/packs`. Each pack optionally has an executable file named `detect` in the root
directory. The intention of this `detect` script is to determine if the pack should be used with
the given application.

A detect executable takes one argument: the directory in which `prow create` was executed. The
executable should be portable across most systems (read: not a Ruby/Python script unless it's
a self-contained binary). The executable should exit with a non-zero exit code if the pack is not
relevant for the app, otherwise it should display the pack's name and exit zero.

For example, here is how a Python pack's detect executable would look like:

```
#!/usr/bin/env bash

APP_DIR=$1

# Exit early if app is clearly not Python.
if [ ! -f $APP_DIR/requirements.txt ] && [ ! -f $APP_DIR/setup.py ]; then
  exit 1
fi

echo Python
```

`prow create` iterates through packs alphabetically with a "first one wins" approach to detection.
If a pack does not include a detect executable, it is considered a "loser". Pack detection can be
overridden with the `--pack` flag. The detect script will not be considered and prow will bootstrap
the app with the pack, no questions asked.

[charts]: https://github.com/kubernetes/helm/blob/master/docs/charts.md
