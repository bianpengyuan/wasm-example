module github.com/bianpengyuan/wasm-example

go 1.13

replace github.com/bianpengyuan/istio-wasm-sdk => /home/bianpengyuan_google_com/workspace/istio-wasm-sdk

require (
	github.com/bazelbuild/bazel-gazelle v0.20.0 // indirect
	github.com/bianpengyuan/istio-wasm-sdk v0.0.0-20200403033522-15a8cdec35ae
	github.com/golang/protobuf v1.3.3
	google.golang.org/grpc v1.27.1
	istio.io/pkg v0.0.0-20200401184616-e9c30fcd524c
)
