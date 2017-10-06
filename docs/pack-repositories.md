# How to create and maintain a pack repository

Pack repositories are external sources of Draft packs. They can be created by anyone to provide their own packs to any Draft user.

## Creating a pack repository

A pack repository is usually a Git repository available online, but you can use anything as long as it’s a protocol that a version control system like git, mercurial or svn understands, or even just a directory with files in it.
The name of the repository doesn't matter, though it is recommended as a best practive to start the repository name with `draft-` as a common convention.

Packs follow the same format as the core’s ones, and can be added under `packs/`. By placing packs inside their own directory, it makes the repository organisation easier to grasp, and top-level files are not mixed with packs.

See [Azure/draft](https://github.com/Azure/draft) for an example of a pack repository with a `packs/` subdirectory.

### Installing

If it’s on GitHub, users can add any of your repositories to their repository list with `draft pack-repo add https://github.com/user/draft-repo`. If it’s hosted outside of GitHub, they have to use `draft pack-repo add <URL>`, where `<URL>` is your version control system's clone URL.

## Maintaining a pack repository

A pack repository is just a Git repository so you don’t have to do anything specific when making modifications, apart from committing and pushing your changes.

### Updating

Once your pack repository is installed, Draft will update it each time a user runs `draft pack-repo update`. Outdated packs will then be upgraded.
