# Changelog

## v0.3.0

### Client

* Create default packs on `draft init` for 6 different languages
* Add basedomain logic for ingress resources in default starter packs
* Ignore temporary files from file-watcher
* Switched from `draft.yaml` to `draft.toml`
* Autogenerate release name on `draft create`

### Server

* Connect to tiller via kubernetes service

### Documentation

* Add example applications for all 6 languages
* Update getting-started documentation to use python example application
* Add Governance model

### Test Infrastructure

* Improved test coverage
* Switched to Jenkins for CI
* Push client binaries to azuredraft storage account

## v0.2.0

### Client

* New command: `draft home`
* New command: `draft init`
* Introduced pack detection into `draft create`
* New option flags on `draft up`: `-f`, `--set`, and `--values`
* Intruduced a default Ingress resource with the default nginx pack
* Introduced `draft.yaml`

### Server

* Initialized connection to Helm on startup rather than at build time
* Bumped Helm to commit 1aee50f

### Documentation

* Introduced the --watch flag in the Getting Started Guide 
* Documented the release process

### Test Infrastructure

* Introduced Drone CI!
  * Canary images are uploaded to docker registry
  * Canary clients are uploaded to S3 for linux-arm, linux-i386, linux-amd64, darwin-amd64, and windows-amd64
  * Release images and clients are uploaded, too!
* Unit tests for the client and server were improved over this release.
* Introduced `hack/docker-make.sh` to run the test suite inside a container 

## v0.1.0

Initial release! :tada:
