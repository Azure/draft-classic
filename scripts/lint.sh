#!/usr/bin/env bash

# Mandatory tests
echo -e "\033[0;31mManadatory Linters: These must pass\033[0m"
gometalinter --vendor --tests --deadline=20s --disable-all \
--enable=gofmt \
--enable=misspell \
--enable=deadcode \
--enable=ineffassign \
--enable=vet \
--exclude=pkg/draft/pack/generated \
$(glide novendor)

mandatory=$?

# Optional tests
echo -e "\033[0;32mOptional Linters: These should pass\033[0m"
gometalinter --vendor --tests --deadline=20s --disable-all \
--enable=golint \
--exclude=pkg/draft/pack/generated \
$(glide novendor)

exit $mandatory
