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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

type LogEntry struct {
	SourceName           string `json:"source_name"`
	SourceNamespace      string `json:"source_namespace"`
	SourceWorkload       string `json:"source_workload"`
	DestinationName      string `json:"destination_name"`
	DestinationNamespace string `json:"destination_namespace"`
	DestinationWorkload  string `json:"destination_workload"`
}

type CheckLog struct {
	N int
	S *LoggingServer
}

var _ framework.Step = &CheckLog{}

func (c *CheckLog) Run(p *framework.Params) error {
	select {
	case req := <-c.S.Req:
		if err := verifyLogEntry(req, c.N); err != nil {
			return err
		}
	case <-time.After(15 * time.Second):
		return errors.New("timeout waiting for log entry")
	}
	return nil
}

func (c *CheckLog) Cleanup() {}

func verifyLogEntry(got string, n int) error {
	var gotLogEntries []*LogEntry
	if err := json.Unmarshal([]byte(got), &gotLogEntries); err != nil {
		return err
	}
	wantedLogs, err := loadExpectedLogEntries(n)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(wantedLogs, gotLogEntries) {
		return fmt.Errorf("got %v, but want %v", gotLogEntries, wantedLogs)
	}
	return nil
}

func loadExpectedLogEntries(n int) ([]*LogEntry, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return []*LogEntry{}, errors.New("failed to find log entry json file")
	}
	filepath := path.Join(path.Dir(filename), "../testdata/log_entry.json")
	logEntryFile, err := os.Open(filepath)
	if err != nil {
		return []*LogEntry{}, err
	}
	defer logEntryFile.Close()

	logBytes, _ := ioutil.ReadAll(logEntryFile)
	var le LogEntry
	json.Unmarshal(logBytes, &le)
	var logEntries []*LogEntry
	for i := 1; i <= n; i++ {
		logEntries = append(logEntries, &le)
	}
	return logEntries, nil
}
