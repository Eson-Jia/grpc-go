/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Binary server is an example server.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/hashicorp/consul/api"
	pb "google.golang.org/grpc/examples/features/proto/echo"
)

const addr = "192.168.1.42:443"

type ecServer struct {
	pb.UnimplementedEchoServer
	addr string
}

func (s *ecServer) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: fmt.Sprintf("%s (from %s)", req.Message, s.addr)}, nil
}

func register() {
	defaultConf := api.DefaultConfig()
	defaultConf.Address = "localhost:8500"
	client, err := api.NewClient(defaultConf)
	if err != nil {
		log.Fatalln("failed in new client", err)
	}
	service := &api.AgentServiceRegistration{
		Name:    "hello_world_server",
		Address: "192.168.1.42",
		Port:    50051,
		Check: &api.AgentServiceCheck{
			Interval: "10s",
			HTTP:     "http://192.168.1.42:50051/health",
		},
	}
	if err := client.Agent().ServiceRegister(service); err != nil {
		log.Fatalln("failed in service register", err)
	}
}

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &ecServer{addr: addr})
	log.Printf("serving on %s\n", addr)
	register()
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		err := http.Serve(lis, nil)
		if err != nil {
			log.Fatalln("failed in health serve")
		}
	}()
	select {}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
