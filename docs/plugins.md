# Using Draft Plugins

A plugin is a tool that can be accessed through the `draft` CLI, but is not part of the built-in Draft codebase.

Draft plugins have the following features:

- They can be added and removed from a Draft installation without impacting the core Draft tool.
- They can be written in any programming language.
- They integrate with Draft, and will show up in `draft help` and other places.

## Installing a Plugin

Plugins are installed using `draft plugin install <name>`. You can pass in the name of a plugin.

## Finding Plugins

Plugins can be discovered through `draft plugin search`.

```console
$ draft plugin search
NAME            REPOSITORY                      VERSION
pack-repo       github.com/draftcreate/plugins  0.4.2
```

Searching for certain plugins can be done by providing mulitple keywords. `draft plugin search` searches
for plugins with similar names using fuzzy string search.

```console
$ draft plugin search rep
NAME            REPOSITORY                      VERSION
pack-repo       github.com/draftcreate/plugins  0.4.2
```

## Writing your own Plugin

See the [Plugin Cookbook](plugin-cookbook.md) and [DEP 005](reference/dep-005.md) for tips and tricks on building and contributing plugins to Draft.
