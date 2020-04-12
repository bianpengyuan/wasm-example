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

package testserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

type LogEntry struct {
	SourceName             string `json:"source_name"`
	SourceNamespace        string `json:"source_namespace"`
	SourceWorkload         string `json:"source_workload"`
	DestinationName        string `json:"destination_name"`
	DestinationNamespace   string `json:"destination_namespace"`
	DestinationWorkload    string `json:"destination_workload"`
	ResponseFlag           string `json:"response_flag"`
	URLPath                string `json:"url_path"`
	DestinationServiceName string `json:"destination_service_name"`
	DestinationServiceHost string `json:"destination_service_host"`
	ResponseCode           int    `json:"response_code"`
	RequestProtocol        string `json:"request_protocol"`
	URLHost                string `json:"url_host"`
	DestinationAddress     string `json:"destination_address"`
}

type CheckLog struct {
	N       int
	S       *LoggingServer
	Timeout time.Duration
}

var _ framework.Step = &CheckLog{}

func (c *CheckLog) Run(p *framework.Params) error {
	select {
	case req := <-c.S.Req:
		fmt.Println(req)
		if err := verifyLogEntry(p, req, c.N); err != nil {
			return err
		}
	case <-time.After(c.Timeout):
		return errors.New("timeout waiting for log entry")
	}
	return nil
}

func (c *CheckLog) Cleanup() {}

func verifyLogEntry(p *framework.Params, got string, n int) error {
	var gotLogEntries []LogEntry
	if err := json.Unmarshal([]byte(got), &gotLogEntries); err != nil {
		return err
	}
	wantLogEntries, err := loadExpectedLogEntries(p, n)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(wantLogEntries, gotLogEntries) {
		return fmt.Errorf("got %+v, but want %+v", gotLogEntries, wantLogEntries)
	}
	return nil
}

func loadExpectedLogEntries(p *framework.Params, n int) ([]LogEntry, error) {
	logBytes := p.LoadTestData("test/logging/testdata/log_entry.json")
	var le LogEntry
	json.Unmarshal([]byte(logBytes), &le)
	var logEntries []LogEntry
	for i := 1; i <= n; i++ {
		logEntries = append(logEntries, le)
	}
	return logEntries, nil
}
