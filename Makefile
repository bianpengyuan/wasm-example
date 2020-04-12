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

.PHONY: test

build: build_opa build_grpc build_logging

test:
	go test ./... -p 1

build_opa:
	bazel build //example/opa:plugin.wasm

test_opa:
	go test ./test/opa/... -p 1

build_grpc:
	bazel build //example/grpc:plugin.wasm

test_grpc:
	go test ./test/grpc/... -p 1

build_logging:
	bazel build //example/logging:plugin.wasm

test_logging:
	go test ./test/logging/... -p 1
