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
	"log"
	"net"
	"net/http"

	pb "github.com/bianpengyuan/wasm-example/test/headermutation/testserver/proto"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	s *grpc.Server
}

// GetHeaderMutation implements header mutation service
func (s *server) GetHeaderMutation(ctx context.Context, in *pb.HeaderMutationRequest) (*pb.HeaderMutationResponse, error) {
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

// RunHeaderMutationServer initilizes a header mutation server.
func RunHeaderMutationServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("start listening on %v \n", port)
	s := grpc.NewServer()
	pb.RegisterHeaderMutationServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// TearDownHeaderMutationServer initilizes a header mutation server.
func TearDownHeaderMutationServer() {
	// TODO
}
