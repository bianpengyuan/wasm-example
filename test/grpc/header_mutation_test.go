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

package grpc

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
	"github.com/bianpengyuan/wasm-example/test/grpc/testserver"
)

func TestHeaderMutation(t *testing.T) {
	var headerMutationTests = []struct {
		name          string
		userCookie    string
		versionHeader string
	}{
		{"v1", "alice", "v1"},
		{"v2", "bob", "v2"},
		{"empty", "", ""},
	}
	for _, tt := range headerMutationTests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := framework.NewTestParams(map[string]string{})

			grpcPort := params.Ports.Max + 1
			params.Vars["ServerHTTPFilters"] = fmt.Sprintf(
				framework.LoadTestData("test/grpc/testdata/resource/grpc_filter.yaml.tmpl"),
				getHeaderMutationPluginWasm(), strconv.Itoa(int(grpcPort)))
			if err != nil {
				t.Fatalf("failed to initialize test params: %v", err)
			}

			var reqHeaders, respHeaders http.Header
			if tt.userCookie != "" {
				reqHeaders = http.Header{"Cookie": []string{fmt.Sprintf("user=%v", tt.userCookie)}}
			}
			if tt.versionHeader != "" {
				respHeaders = http.Header{"Version": []string{tt.versionHeader}}
			}
			if err := (&framework.Scenario{
				Steps: []framework.Step{
					&framework.XDS{},
					&testserver.Server{Port: grpcPort},
					&framework.ClientServerEnvoy{},
					&framework.Sleep{Duration: 3 * time.Second},
					&framework.HTTPClient{
						Op:              framework.GET,
						URL:             fmt.Sprintf("http://127.0.0.1:%d/echo", params.Ports.ClientPort),
						ReqHeaders:      reqHeaders,
						WantRespCode:    200,
						WantRespHeaders: respHeaders,
					},
				}}).Run(params); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func getHeaderMutationPluginWasm() string {
	workspace, _ := exec.Command("bazel", "info", "workspace").Output()
	return filepath.Join(strings.TrimSuffix(string(workspace), "\n"), "bazel-bin/example/grpc/plugin.wasm")
}
