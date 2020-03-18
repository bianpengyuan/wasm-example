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

package headermutation

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
	"github.com/bianpengyuan/wasm-example/test/headermutation/testserver"
)

func TestHeaderMutation(t *testing.T) {
	ec := getTestEnvoyConfig()
	go testserver.RunHeaderMutationServer()
	framework.NewTest(ec, t).Run(func(ports *framework.Ports) {
		code, header, _, err := framework.HTTPGet(fmt.Sprintf("http://127.0.0.1:%d/echo", ports.AppToClientProxyPort),
			map[string][]string{"Cookie": []string{"user=alice"}})
		fmt.Println(header)
		if err != nil || code != 200 {
			t.Errorf("Failed in request: %v or response code is not expected: %v", err, code)
		}
	})
}

func getTestEnvoyConfig() framework.TestEnvoyConfig {
	var config framework.TestEnvoyConfig
	// Inject server side filter which
	headerMutationFilter := `- name: envoy.filters.http.wasm
  config:
    config:
      vm_config:
        vm_id: "header_mutation_vm"
        runtime: "envoy.wasm.runtime.v8"
        code:
          local: { filename: %v }
      configuration: >-
        {
          "header_mutation_service": "127.0.0.1:50051",
        }`
	config.FiltersBeforeEnvoyRouterInProxyToServer = fmt.Sprintf(headerMutationFilter, getHeaderMutationPluginWasm())

	return config
}

func getHeaderMutationPluginWasm() string {
	workspace, _ := exec.Command("bazel", "info", "workspace").Output()
	return filepath.Join(strings.TrimSuffix(string(workspace), "\n"), "bazel-bin/example/header_mutation/plugin.wasm")
}
