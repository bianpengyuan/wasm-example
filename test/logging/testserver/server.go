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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
)

type LoggingServer struct {
	Req  chan string
	Port uint16
}

var _ framework.Step = &LoggingServer{}

func (l *LoggingServer) Run(p *framework.Params) error {
	l.Req = make(chan string, 10)
	go func() {
		http.HandleFunc("/", l.handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", l.Port), nil))
	}()
	return nil
}

func (l *LoggingServer) Cleanup() {}

func (l *LoggingServer) handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	l.Req <- string(body)
}
