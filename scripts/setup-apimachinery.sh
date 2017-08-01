#!/usr/bin/env bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Copies the current versions of apimachinery and client-go out of the
# main kubernetes repo.  These repos are currently out of sync and not
# versioned.
set -euo pipefail


rm -rf ./vendor/k8s.io/{kube-aggregator,apiserver,apimachinery,client-go}

cp -r ./vendor/k8s.io/kubernetes/staging/src/k8s.io/{kube-aggregator,apiserver,apimachinery,client-go} ./vendor/k8s.io

rm -rf ./vendor/github.com/generaltso/linguist
cd vendor/github.com/generaltso/
git clone https://github.com/generaltso/linguist
cd linguist/
git checkout 9202505feb6930ac6945b5005715d2b03853ac9c
git submodule update --init
go generate
go generate ./data
