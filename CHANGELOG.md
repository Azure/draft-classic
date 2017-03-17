# Changelog

## v0.2.0

### Client

* New command: `prow home` [#124](https://github.com/deis/prow/pull/124)
* New command: `prow init` [#126](https://github.com/deis/prow/pull/126)
* Introduced pack detection into `prow create` [#138](https://github.com/deis/prow/pull/138)
* New option flags on `prow up`: `-f`, `--set`, and `--values` [#139](https://github.com/deis/prow/pull/139)
* Intruduced a default Ingress resource with the default nginx pack [#151](https://github.com/deis/prow/pull/151)

### Server

* Initialized connection to Helm on startup rather than at build time [#106](https://github.com/deis/prow/pull/106)
* Bumped Helm to commit 1aee50f [#155](https://github.com/deis/prow/pull/155)

### Documentation

* Introduced the --wait flag in the Getting Started Guide [#108](https://github.com/deis/prow/pull/108)
* Documented the release process [#143](https://github.com/deis/prow/pull/143)

### Test Infrastructure

* Introduced Drone CI! [#128](https://github.com/deis/prow/pull/128)
  * Canary images are uploaded to quay.io/deis/prow
  * Canary clients are uploaded to S3 for linux-i386, linux-amd64, darwin-amd64, and windows-amd64
  * Release images and clients are uploaded, too!
* Unit tests for the client and server were improved over this release. [#172](https://github.com/deis/prow/pull/172) [#173](https://github.com/deis/prow/pull/173)
* Introduced `hack/docker-make.sh` to run test suite inside a container [#177](https://github.com/deis/prow/pull/177)

## v0.1.0

Initial release! :tada:
