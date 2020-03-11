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
	"testing"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

func TestOPAPluginAllow(t *testing.T) {
  	var opa OpaServer
	if err := opa.SetupOpaServer("testdata/rule/opa_rule.rego"); err != nil {
		t.Fatalf("fail to initialize test OPA server: %v", err)
	}
	defer opa.TearDownOpaServer(t)
	ec := getTestEnvoyConfig()
	framework.NewTest(ec, t).Run(func(ports *framework.Ports) {
		if code, _, err := framework.HTTPGet(fmt.Sprintf("http://127.0.0.1:%d/echo", ports.AppToClientProxyPort)); err != nil || code != 200 {
			t.Errorf("Failed in request: %v or response code is not expected: %v", err, code)
		}
	})
}

func TestOPAPluginDeny(t *testing.T) {
	var opa OpaServer
	if err := opa.SetupOpaServer("testdata/rule/opa_rule.rego"); err != nil {
		t.Fatalf("fail to initialize test OPA server: %v", err)
	}
	defer opa.TearDownOpaServer(t)
	ec := getTestEnvoyConfig()
	framework.NewTest(ec, t).Run(func(ports *framework.Ports) {
		if code, _, err := framework.HTTPPost(fmt.Sprintf("http://127.0.0.1:%d/echo", ports.AppToClientProxyPort), "", ""); err != nil || code != 403 {
			t.Errorf("Failed in request: %v or response code is not expected: %v", err, code)
		}
	})
}

