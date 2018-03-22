#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
DRAFT_ROOT="${BASH_SOURCE[0]%/*}/.."

cd "$DRAFT_ROOT"

run_unit_test() {
  echo "Running unit tests"
  go get -u github.com/jstemmer/go-junit-report
  mkdir -p /tmp/test-results
  trap "go-junit-report </tmp/test-results/go-test.out > /tmp/test-results/go-test-report.xml" EXIT
  make test-unit | tee /tmp/test-results/go-test.out
}

run_style_check() {
  echo "Running style checks"
  make test-lint
}

run_unit_test
run_style_check
