## Copyright 2020 Istio Authors
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
##     http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.

all: build_opa test_opa build_header_mutation test_header_mutation

build_opa:
	bazel build //example/opa:plugin.wasm

test_opa:
	bazel build //example/opa:plugin.wasm && go test ./test/opa/... -count=1

build_header_mutation:
	bazel build //example/header_mutation:plugin.wasm

test_header_mutation:
	bazel build //example/header_mutation:plugin.wasm && go test ./test/headermutation/... -count=1
