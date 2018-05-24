# Plugin Cookbook

Plugins are a package definition written in Lua. It can be created with `draft plugin create <name>`. Plugins use the Lua runtime to provide simple scripting capabilities that other markdown languages cannot provide on their own.

## Draft terminology

| Term       | Description                         | Example                                                                            |
|------------|-------------------------------------|------------------------------------------------------------------------------------|
| Plugin     | The package definition              | ~/.draft/plugins/repositories/github.com/draftcreate/plugins/Plugins/pack-repo.lua |
| Repository | A git repository containing plugins | ~/.draft/plugins/repositories/github.com/draftcreate/plugins                       |
| Installed  | All plugins are installed here      | ~/.draft/plugins/installed/pack-repo/0.4.2/                                        |

## An introduction

Draft uses git for contributing to the project.

Draft installs plugins to `$(draft home)/plugins/installed`. We suggest you use `find` on a few of the plugins in `draft home` to see how it is all arranged.

Packages are installed according to their Plugins, which are written in Lua and live in `$(draft home)/plugins`.

## Basic instructions

Make sure you run `draft plugin update` before you start. This prepares your repositories by bumping them to the latest revision.

Before submitting a new plugins to core, make sure your package:

- isn’t already in core plugins
- isn’t already waiting to be merged (check the pull request queue)
- is still supported by upstream
- has a stable, tagged version (i.e. not just a work-in-progress with no versions)

## Create the Plugins

```
draft plugin update # make sure we've got a fresh checkout of master
vim $(draft plugin create foo)
```

## Test the Plugins

...By installing it!

```
draft plugin install foo
```

## Commit

Everything is built on git, so contribution is easy:

```
vim $(draft plugin create foo)
cd $(draft home)/plugins/repositories/github.com/draftcreate/plugins
# Create a new git branch for your plugins so your pull request is easy to
# modify if any changes come up during review.
git checkout -b <some-descriptive-name>
git add Plugins/foo.lua
git commit
```

## Fork

Now you just need to push your commit to GitHub.

If you haven’t forked Draft's plugin core yet, go to the draftcreate/plugins repository and hit the Fork button.

If you have already forked Draft's plugin core on GitHub, then you can manually push (just make sure you have been pulling from the draftcreate/plugins master):

```
git push https://github.com/user/draft-plugins/ <what-you-called-your-branch>
```

Now, open a pull request for your changes.

- One plugin per commit; one commit per plugin
- Keep merge commits out of the pull request
