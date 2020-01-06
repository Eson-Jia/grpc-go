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
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/examples/features/name_resolving/client/resolver"
	"log"
	"net"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"

	"github.com/hashicorp/consul/api"
	pb "google.golang.org/grpc/examples/features/proto/echo"
)

const host = "192.168.1.42"
const port = 50051

var addr = fmt.Sprintf("%s:%d", host, port)

type ecServer struct {
	pb.UnimplementedEchoServer
	addr string
}

func (s *ecServer) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: fmt.Sprintf("%s (from %s)", req.Message, s.addr)}, nil
}

func registerConsul() {
	defaultConf := api.DefaultConfig()
	defaultConf.Address = "localhost:8500"
	client, err := api.NewClient(defaultConf)
	if err != nil {
		log.Fatalln("failed in new client", err)
	}
	service := &api.AgentServiceRegistration{
		Name:    "Greet",
		Address: host,
		Port:    port,
		Weights: &api.AgentWeights{
			Passing: 5,
			Warning: 5,
		},
	}
	if err := client.Agent().ServiceRegister(service); err != nil {
		log.Fatalln("failed in service register", err)
	}
}

func registerEtcd() error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"localhost:2379",
		},
	})
	if err != nil {
		log.Fatalln("failed in new client:", err)
	}
	leaseRep, err := client.Grant(context.TODO(), 10)
	if err != nil {
		return err
	}
	kpaChan, err := client.KeepAlive(context.Background(), leaseRep.ID)
	if err != nil {
		return err
	}
	go func() {
		for resp := range kpaChan {
			log.Println(resp)
		}
	}()
	config := resolver.Node{
		Host:   "192.168.1.42",
		Port:   50051,
		Weight: 10,
		ID:     "1",
	}
	buff, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = client.Put(context.Background(), fmt.Sprintf("/etcd/service/greet/%s", config.ID), string(buff), clientv3.WithLease(leaseRep.ID))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &ecServer{addr: addr})
	log.Printf("serving on %s\n", addr)
	if false {
		registerConsul()
	}
	if true {
		registerEtcd()
	}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
