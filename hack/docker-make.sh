#!/usr/bin/env bash

cd $(dirname $0)

docker run \
    -v $PWD/..:/go/src/github.com/deis/draft \
    --workdir /go/src/github.com/deis/draft \
    deis/go-dev:v0.22.0 \
    make $@
