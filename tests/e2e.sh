#!/usr/bin/env bash
#
# Usage: ./e2e.sh
#
# This script assumes an existing single node k8s cluster with helm and draftd installed. It
# installs all the apps under testdata and checks that they pass or fail, depending on their
# success condition.

cd $(dirname $0)
PATH="$(pwd)/../bin:$PATH"

echo "testing apps that are expected to pass"
pushd testdata/good > /dev/null
for app in */; do
    echo "switching to ${app}"
    pushd "${app}" > /dev/null
    # strip trailing forward slash
    app=${app%/}
    draft up -e nowatch
    echo "checking that ${app} v1 was released"
    revision=$(helm list | grep "${app}" | awk '{print $2}')
    name=$(helm list | grep "${app}" | awk '{print $1}')
    if [[ "$revision" != "1" ]]; then
        echo "Expected REVISION == 1, got $revision"
        exit 1
    fi
    echo "GOOD"
    # deploy the app again and check that the update is seen upstream
    draft up -e nowatch
    echo "checking that ${app} v2 was released"
    revision=$(helm list | grep "${app}" | awk '{print $2}')
    if [[ "$revision" != "2" ]]; then
        echo "Expected REVISION == 2, got $revision"
        exit 1
    fi
    echo "GOOD"
    echo "deleting the helm release for ${app}: ${name}"
    # clean up
    helm delete --purge "${name}"
    echo "GOOD"
    popd > /dev/null
done
popd > /dev/null

echo "testing apps that are expected to fail"
pushd testdata/bad > /dev/null
for app in */; do
    echo "switching to ${app}"
    pushd "${app}" > /dev/null
    # strip trailing forward slash
    app=${app%/}
    draft up
    echo "checking that ${app} v1 was NOT released"
    release=$(helm list | grep "${app}")
    if [[ "$release" != "" ]]; then
        echo "Expected no release to exist , got $release"
        exit 1
    fi
    echo "GOOD"
    popd > /dev/null
done
popd > /dev/null

echo "e2e tests finished."
