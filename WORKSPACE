workspace(name = "wasm_example")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "istio_wasm_sdk",
    strip_prefix = "istio-wasm-sdk-a66b0afd52302962923f3477fe2f58fbb4a3c22f",
    url = "https://github.com/bianpengyuan/istio-wasm-sdk/archive/a66b0afd52302962923f3477fe2f58fbb4a3c22f.tar.gz",
)

load("@istio_wasm_sdk//bazel:sdk_dependencies.bzl", "sdk_dependencies")

sdk_dependencies()

load("@istio_wasm_sdk//bazel:sdk_dependency_imports.bzl", "sdk_dependency_imports")

sdk_dependency_imports()
