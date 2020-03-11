// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opa

import (
  "fmt"
	"go/build"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"
  "testing"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

type OpaServer struct {
  opaProcess *os.Process
}

func downloadOpaServer() (string, error) {
  outputPath := fmt.Sprintf("%s/out/%s_%s", build.Default.GOPATH, runtime.GOOS, runtime.GOARCH)
  dst := fmt.Sprintf("%v/opa", outputPath)
	if _, err := os.Stat(dst); err == nil {
	  return dst, nil
	}
  opaURL := "https://openpolicyagent.org/downloads/latest/opa_linux_amd64"
  fmt.Printf("download opa server to %v from %v", dst, opaURL)
	donwloadCmd := exec.Command("bash", "-c", fmt.Sprintf("curl -L -o %v %v", dst, opaURL))
	donwloadCmd.Stderr = os.Stderr
	donwloadCmd.Stdout = os.Stdout
	err := donwloadCmd.Run()
	if err != nil {
		return "", fmt.Errorf("fail to run opa download command: %v", err)
  }
  chmodCmd := exec.Command("bash", "-c", fmt.Sprintf("chmod 755 %v", dst))
  chmodCmd.Stderr = os.Stderr
	chmodCmd.Stdout = os.Stdout
	err = chmodCmd.Run()
	if err != nil {
		return "", fmt.Errorf("fail to chmod for opa: %v", err)
  }
	return dst, nil
}

func (s *OpaServer) SetupOpaServer(ruleFilePath string) error {
  // Download Opa Server
  opaPath, err := downloadOpaServer()
  if err != nil {
    return err
  }

  // Run Opa Server with given rule file
  opaServerCmd := fmt.Sprintf("%v run --server %v", opaPath, ruleFilePath)
  fmt.Printf("start opa server: %v", opaServerCmd)
  opaCmd := exec.Command("bash", "-c", opaServerCmd)
	opaCmd.Stderr = os.Stderr
	opaCmd.Stdout = os.Stdout
	err = opaCmd.Start()
  if err != nil {
    return err
  }
  s.opaProcess = opaCmd.Process
	return nil
}


func (s *OpaServer) TearDownOpaServer(t *testing.T) {
  err := s.opaProcess.Kill()
  if err != nil {
    t.Errorf("failed to kill opa server: %v", err)
  }
}

func getOpaPluginWasm() string {
	workspace, _ := exec.Command("bazel", "info", "workspace").Output()
	return filepath.Join(strings.TrimSuffix(string(workspace), "\n"), "bazel-bin/example/opa/plugin.wasm")
}

func getTestEnvoyConfig() framework.TestEnvoyConfig {
	var config framework.TestEnvoyConfig
  // Inject server side filter which
  opaFilter := `
- name: envoy.filters.http.wasm
  config:
    config:
      vm_config:
        vm_id: "opa_vm"
        runtime: "envoy.wasm.runtime.v8"
        code:
          local: { filename: %v }
      configuration: >-
        {
          "opa_cluster_name": "opa_policy_server",
          "opa_service_host": "localhost:8181",
          "fail_open": "false",
        }`
	config.FiltersBeforeEnvoyRouterInProxyToServer = fmt.Sprintf(opaFilter, getOpaPluginWasm())

	// Add a cluster for OPA server
  config.ServerEnvoyExtraCluster = `- name: opa_policy_server
  connect_timeout: 5s
  type: STATIC
  load_assignment:
    cluster_name: opa_policy_server
    endpoints:
    - lb_endpoints:
      - endpoint:
          address:
            socket_address:
              address: 127.0.0.1
              port_value: 8181`

  // Enable mtls
  config.ClientClusterTLSContext = `transport_socket:
  name: envoy.transport_sockets.tls
  typed_config:
    "@type": type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
    common_tls_context:
      tls_certificates:
      - certificate_chain: { filename: "testdata/certs/client.cert" }
        private_key: { filename: "testdata/certs/client-key.cert" }
      validation_context:
        trusted_ca: { filename: "testdata/certs/root.cert" }
    sni: server.com`

  config.ServerTLSContext = `transport_socket:
  name: envoy.transport_sockets.tls
  typed_config:
    "@type": type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext
    common_tls_context:
      tls_certificates:
      - certificate_chain: { filename: "testdata/certs/server.cert" }
        private_key: { filename: "testdata/certs/server-key.cert" }
      validation_context:
        trusted_ca: { filename: "testdata/certs/root.cert" }
    require_client_certificate: true`

	return config
}
