#!/usr/bin/env bash

cd $(dirname $0)

docker run \
    -v $PWD/..:/go/src/github.com/deis/prow \
    --workdir /go/src/github.com/deis/prow \
    deis/go-dev:v0.22.0 \
    make $@
