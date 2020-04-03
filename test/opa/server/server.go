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

package server

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"runtime"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

// OpaServer ...
type OpaServer struct {
	opaProcess   *os.Process
	RuleFilePath string
}

var _ framework.Step = &OpaServer{}

// Run ...
func (o *OpaServer) Run(p *framework.Params) error {
	opaPath, err := downloadOpaServer()
	if err != nil {
		return err
	}

	// Run Opa Server with given rule file
	opaServerCmd := fmt.Sprintf("%v run --server %v", opaPath, o.RuleFilePath)
	fmt.Printf("start opa server: %v", opaServerCmd)
	opaCmd := exec.Command("bash", "-c", opaServerCmd)
	opaCmd.Stderr = os.Stderr
	opaCmd.Stdout = os.Stdout
	err = opaCmd.Start()
	if err != nil {
		return err
	}
	o.opaProcess = opaCmd.Process
	return nil
}

// Cleanup ...
func (o *OpaServer) Cleanup() {
	o.opaProcess.Kill()
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
