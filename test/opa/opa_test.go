// Copyright 2020 Istio Authors
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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
	opa "github.com/bianpengyuan/wasm-example/test/opa/server"
)

func TestOPAPlugin(t *testing.T) {
	var tests = []struct {
		name         string
		op           framework.HTTPOperation
		cacheHit     int
		cacheMiss    int
		requestCount int
		delay        time.Duration
		wantRespCode int
	}{
		{"allow", framework.GET, 9, 1, 10, 0, 200},
		{"deny", framework.POST, 9, 1, 10, 0, 403},
		{"cache_expire", framework.POST, 2, 2, 4, 4 * time.Second, 403},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := framework.NewTestParams(map[string]string{
				"ClientTLSContext":    framework.LoadTestData("test/opa/testdata/transport_socket/client_tls_context.yaml.tmpl"),
				"ServerTLSContext":    framework.LoadTestData("test/opa/testdata/transport_socket/server_tls_context.yaml.tmpl"),
				"ServerStaticCluster": framework.LoadTestData("test/opa/testdata/resource/opa_cluster.yaml.tmpl"),
				"OpaPluginFilePath":   getOpaPluginWasm(),
				"CacheHit":            strconv.Itoa(tt.cacheHit),
				"CacheMiss":           strconv.Itoa(tt.cacheMiss),
			})
			params.Vars["ServerHTTPFilters"] = params.LoadTestData("test/opa/testdata/resource/opa_filter.yaml.tmpl")
			if err != nil {
				t.Fatalf("failed to initialize test params: %v", err)
			}

			if err := (&framework.Scenario{
				Steps: []framework.Step{
					&framework.XDS{},
					&opa.OpaServer{RuleFilePath: "testdata/rule/opa_rule.rego"},
					&framework.ClientServerEnvoy{},
					&framework.Repeat{
						N: tt.requestCount,
						Step: &framework.Scenario{
							Steps: []framework.Step{
								&framework.HTTPClient{
									Op:           tt.op,
									URL:          fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
									WantRespCode: tt.wantRespCode,
								},
								&framework.Sleep{Duration: tt.delay},
							},
						},
					},
					&framework.Stats{
						AdminPort: params.Ports.ServerAdminPort,
						Matchers: map[string]framework.StatMatcher{
							"envoy_policy_cache_count": &framework.ExactStat{Metric: "test/opa/testdata/stats/cache.yaml.tmpl"},
						}},
				}}).Run(params); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOPAPluginReload(t *testing.T) {
	params, err := framework.NewTestParams(map[string]string{
		"ClientTLSContext":    framework.LoadTestData("test/opa/testdata/transport_socket/client_tls_context.yaml.tmpl"),
		"ServerTLSContext":    framework.LoadTestData("test/opa/testdata/transport_socket/server_tls_context.yaml.tmpl"),
		"OpaPluginFilePath":   getOpaPluginWasm(),
		"ServerStaticCluster": framework.LoadTestData("test/opa/testdata/resource/opa_cluster.yaml.tmpl"),
		"CacheHit":            strconv.Itoa(19),
		"CacheMiss":           strconv.Itoa(1),
	})
	params.Vars["ServerHTTPFilters"] = params.LoadTestData("test/opa/testdata/resource/opa_filter.yaml.tmpl")
	if err != nil {
		t.Fatalf("failed to initialize test params: %v", err)
	}

	if err := (&framework.Scenario{
		Steps: []framework.Step{
			&framework.XDS{},
			&opa.OpaServer{RuleFilePath: "testdata/rule/opa_rule.rego"},
			&framework.ClientServerEnvoy{},
			&framework.Sleep{Duration: 3 * time.Second},
			&framework.Repeat{
				N: 10,
				Step: &framework.HTTPClient{
					Op:           framework.POST,
					URL:          fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
					WantRespCode: 403,
				},
			},
			&framework.Update{Node: "server", Version: "1"},
			&framework.Repeat{
				N: 10,
				Step: &framework.HTTPClient{
					Op:           framework.POST,
					URL:          fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
					WantRespCode: 403,
				},
			},
			&framework.Stats{
				AdminPort: params.Ports.ServerAdminPort,
				Matchers: map[string]framework.StatMatcher{
					"envoy_policy_cache_count": &framework.ExactStat{Metric: "test/opa/testdata/stats/cache.yaml.tmpl"},
				}},
		}}).Run(params); err != nil {
		t.Fatal(err)
	}
}

func getOpaPluginWasm() string {
	workspace, _ := exec.Command("bazel", "info", "workspace").Output()
	return filepath.Join(strings.TrimSuffix(string(workspace), "\n"), "bazel-bin/example/opa/plugin.wasm")
}
