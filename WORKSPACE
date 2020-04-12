workspace(name = "wasm_example")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "istio_wasm_sdk",
    strip_prefix = "istio-wasm-sdk-d2b3614714ae829cdbbcdb2e5fc60d0b010ec862",
    url = "https://github.com/bianpengyuan/istio-wasm-sdk/archive/d2b3614714ae829cdbbcdb2e5fc60d0b010ec862.tar.gz",
)

load("@istio_wasm_sdk//bazel:sdk_dependencies.bzl", "sdk_dependencies")

sdk_dependencies()

load("@istio_wasm_sdk//bazel:sdk_dependency_imports.bzl", "sdk_dependency_imports")

sdk_dependency_imports()

http_archive(
    name = "com_github_tencent_rapidjson",
    build_file = "@wasm_example//bazel/external:rapidjson.BUILD",
    strip_prefix = "rapidjson-1.1.0",
    url = "https://github.com/Tencent/rapidjson/archive/v1.1.0.tar.gz",
)
