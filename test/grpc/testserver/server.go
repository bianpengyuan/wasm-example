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
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/bianpengyuan/istio-wasm-sdk/istio/test/framework"
	pb "github.com/bianpengyuan/wasm-example/test/grpc/testserver/proto"
	"google.golang.org/grpc"
	"istio.io/pkg/log"
)

type Server struct {
	Port uint16
	s    *grpc.Server
}

var _ framework.Step = &Server{}

func (srv *Server) runHeaderMutationServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", srv.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("start listening on %v \n", srv.Port)
	srv.s = grpc.NewServer()
	pb.RegisterHeaderMutationServiceServer(srv.s, srv)
	if err := srv.s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (srv *Server) Run(p *framework.Params) error {
	go srv.runHeaderMutationServer()
	return nil
}

func (srv *Server) Cleanup() {
	srv.s.Stop()
}

// GetHeaderMutation implements header mutation service
func (srv *Server) GetHeaderMutation(ctx context.Context, in *pb.HeaderMutationRequest) (*pb.HeaderMutationResponse, error) {
	fmt.Println("received get header mutation requests")
	header := http.Header{}
	header.Add("Cookie", in.GetCookie())
	request := http.Request{Header: header}
	var resp pb.HeaderMutationResponse
	resp.HeaderMutation = make(map[string]string)
	for _, c := range request.Cookies() {
		if c.Name != "user" {
			continue
		}
		if c.Value == "alice" {
			resp.HeaderMutation["version"] = "v1"
		} else if c.Value == "bob" {
			resp.HeaderMutation["version"] = "v2"
		}
	}
	return &resp, nil
}
