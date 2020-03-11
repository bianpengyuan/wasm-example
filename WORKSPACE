workspace(name = "wasm_example")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "istio_wasm_sdk",
    strip_prefix = "istio-wasm-sdk-5533b0f7a2f3f35cb8b9fd93ac97bd8581eda8e8",
    url = "https://github.com/bianpengyuan/istio-wasm-sdk/archive/5533b0f7a2f3f35cb8b9fd93ac97bd8581eda8e8.tar.gz",
)

load("@istio_wasm_sdk//bazel:sdk_dependencies.bzl", "sdk_dependencies")

sdk_dependencies()

load("@istio_wasm_sdk//bazel:sdk_dependency_imports.bzl", "sdk_dependency_imports")

sdk_dependency_imports()
