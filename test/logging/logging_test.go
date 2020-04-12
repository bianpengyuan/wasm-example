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

package logging

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
	"github.com/bianpengyuan/wasm-example/test/logging/testserver"
)

func TestLogger(t *testing.T) {
	params, err := framework.NewTestParams(map[string]string{
		"LoggingPluginFilePath": getLoggingPluginWasm(),
	})
	if err != nil {
		t.Fatalf("failed to initialize test params: %v", err)
	}

	params.Vars["ServerHTTPFilters"] = params.LoadTestData("test/logging/testdata/logging_filter.yaml.tmpl")
	params.Vars["ServerStaticCluster"] = params.LoadTestData("test/logging/testdata/logging_cluster.yaml.tmpl")
	ls := &testserver.LoggingServer{Port: params.Ports.Max + 1}
	if err := (&framework.Scenario{
		Steps: []framework.Step{
			&framework.XDS{},
			ls,
			&framework.ClientServerEnvoy{},
			&framework.Repeat{
				N: 10,
				Step: &framework.HTTPClient{
					Op:           framework.GET,
					URL:          fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
					WantRespCode: 200,
				},
			},
			&testserver.CheckLog{N: 10, S: ls, Timeout: 12 * time.Second},
		}}).Run(params); err != nil {
		t.Fatal(err)
	}
}

func TestLoggerVMReload(t *testing.T) {
	params, err := framework.NewTestParams(map[string]string{
		"LoggingPluginFilePath": getLoggingPluginWasm(),
		"VMNameSuffix":          "_0",
	})
	if err != nil {
		t.Fatalf("failed to initialize test params: %v", err)
	}

	params.Vars["ServerHTTPFilters"] = params.LoadTestData("test/logging/testdata/logging_filter.yaml.tmpl")
	params.Vars["ServerStaticCluster"] = params.LoadTestData("test/logging/testdata/logging_cluster.yaml.tmpl")
	ls := &testserver.LoggingServer{Port: params.Ports.Max + 1}
	if err := (&framework.Scenario{
		Steps: []framework.Step{
			&framework.XDS{},
			ls,
			&framework.ClientServerEnvoy{},
			&framework.Repeat{
				N: 10,
				Step: &framework.HTTPClient{
					Op:           framework.GET,
					URL:          fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
					WantRespCode: 200,
				},
			},
			&framework.UpdateParamsVars{
				Vars: map[string]string{
					"VMNameSuffix": "_1",
				},
			},
			&framework.ReloadParamsVars{
				Vars: map[string]string{
					"ServerHTTPFilters": "test/logging/testdata/logging_filter.yaml.tmpl",
				},
			},
			&framework.Update{Node: "server", Version: "1"},
			&framework.Sleep{Duration: 3 * time.Second},
			&testserver.CheckLog{N: 10, S: ls, Timeout: 3 * time.Second},
		}}).Run(params); err != nil {
		t.Fatal(err)
	}
}

func getLoggingPluginWasm() string {
	workspace, _ := exec.Command("bazel", "info", "workspace").Output()
	return filepath.Join(strings.TrimSuffix(string(workspace), "\n"), "bazel-bin/example/logging/plugin.wasm")
}
