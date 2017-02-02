#!/usr/bin/env bash
# Usage: ./e2e.sh
#
# This script assumes an existing single node k8s cluster with helm and prowd installed. This script assumes
# that prowd is available at http://k8s.cluster:44135
# To do this, make a new rule in /etc/hosts
#
# $ cat /etc/hosts | grep k8s.cluster
# 10.245.1.3      k8s.cluster
#
# Port 44135 is the port in which prowd is exposed on the host. This will eventually be fixed with
# https://github.com/deis/prow/issues/1

# fail fast
set -eof pipefail

CONTAINER_TIMEOUT=${CONTAINER_TIMEOUT:-1m}

cd testdata/example-dockerfile-http

# TODO(bacongobbler): replace this with `prow up` once it is readily available
# deploy the app
prow up

# wait for the container to come up
sleep $CONTAINER_TIMEOUT

revision=$(helm list | grep example-dockerfile-http | awk '{print $2}')
if [[ "$revision" != "1" ]]; then
	echo "Expected REVISION == 1, got '$revision'"
	exit 1
fi

# example-dockerfile-http exposes itself on port 44144
app_output=$(curl -sS http://k8s.cluster:44144)
if [[ "$app_output" != "Powered by Prow" ]]; then
	echo "Expected 'Powered by Prow', got '$app_output'"
	exit 1
fi

# TODO(bacongobbler): replace this with `prow up` once it is readily available
# deploy the app again, changing POWERED_BY and check that the update is seen upstream
echo "ENV POWERED_BY Kubernetes" >> Dockerfile
prow up

sleep $CONTAINER_TIMEOUT
app_output_2=$(curl -sS http://k8s.cluster:44144)
if [[ "$app_output_2" != "Powered by Kubernetes" ]]; then
	echo "Expected 'Powered by Kubernetes', got '$app_output_2'"
	exit 1
fi
