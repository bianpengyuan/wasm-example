workspace(name = "wasm_example")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "istio_wasm_sdk",
    strip_prefix = "istio-wasm-sdk-e8e7b859f748e6ca2872606daf153aa9791e9e8a",
    url = "https://github.com/bianpengyuan/istio-wasm-sdk/archive/e8e7b859f748e6ca2872606daf153aa9791e9e8a.tar.gz",
)

load("@istio_wasm_sdk//bazel:sdk_dependencies.bzl", "sdk_dependencies")

sdk_dependencies()

load("@istio_wasm_sdk//bazel:sdk_dependency_imports.bzl", "sdk_dependency_imports")

sdk_dependency_imports()
