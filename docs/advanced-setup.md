# Advanced Setup

Once you have `draft` in your PATH, you'll want to run the `$ draft init`. `$ draft init` configures your local machine to work with Draft. It installs a set of default draft packs, plugins, and other directories required to work with Draft.

Use the `--dry-run` flag to see what what Draft will install without actually doing anything: `$ draft init --dry-run`


Your team may also have a set of plugins and pack repositories (repos) that they want to use or you all may even want to override some of the default ones that are installed. In this scenario, you can pass a TOML file in via the `--config` flag with the specified plugins and pack repositories you'd like to install during the `draft init` process.

Specify in double brackets what you'd like to install: `[[plugin]]` or `[[repo]]`. Under that, specifiy some key/value pairs for information on the thing you're trying to install. Specify a `name` for the plugin or pack repository (repo) to install along with a `url` (location of the plugin or repo). Specify what version of a plugin you'd like to install using the `version` key. Case matters here. Keys need to be lowercase. See example below.

In `config-file.toml`:

```toml
[[plugin]]
name = "plugin1"
url = "url to plugin1"
version = "1.0.0"

[[plugin]]
name = "plugin2"
url = "url to plugin2"
version = "1.0.0"

[[repo]]
name = "pack-repo"
url = "url to pack repository"

[[repo]]
name = "another-repo"
url = "url to another pack repository"
```

On the command line:
```console
$ draft init --config config-file.toml
```

If a plugin or pack repository with the same name as the one you have specified already exists, the existing plugin or pack repository will be deleted and replaced with the one you've specified in the toml file passed in.
