
# Copyright 2020 Istio Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#############################################################################

cc_binary(
    name = "plugin.wasm",
    srcs = [
        "plugin.cc",
        "plugin.h",
    ],
    additional_linker_inputs = ["@proxy_wasm_cpp_sdk//:jslib"],
    linkopts = [
        "--js-library",
        "proxy_wasm_intrinsics.js",
        "-UHAVE_PTHREAD",
    ],
    deps = [
        ":config_cc_proto",
        ":opa_payload_cc_proto",
    ],
)

cc_proto_library(
    name = "opa_payload_cc_proto",
    deps = [":opa_payload_proto"],
)

proto_library(
    name = "opa_payload_proto",
    srcs = ["opa_payload.proto"],
)


cc_proto_library(
    name = "config_cc_proto",
    deps = [":config_proto"],
)

proto_library(
    name = "config_proto",
    srcs = ["config.proto"],
    deps = [
        "@com_google_protobuf//:duration_proto",
    ],
)
