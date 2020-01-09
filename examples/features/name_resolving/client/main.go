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

// Binary client is an example client.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc/balancer/roundrobin"

	"google.golang.org/grpc"

	_ "google.golang.org/grpc/examples/features/name_resolving/client/resolver/consul_dns"
	_ "google.golang.org/grpc/examples/features/name_resolving/client/resolver/custom"
	// 发现导入此插件以后发现设置环境变量:
	// GRPC_GO_LOG_VERBOSITY_LEVEL=99
	// GRPC_GO_LOG_SEVERITY_LEVEL=info
	// 不能开启测试日志打印
	_ "google.golang.org/grpc/examples/features/name_resolving/client/resolver/etcd"
	ecpb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/resolver"
)

const (
	exampleScheme      = "example"
	exampleServiceName = "resolver.example.grpc.io"
	customScheme       = "custom"
	customServiceName  = "resolver.custom"
	etcdServiceName    = "etcd/service/greet"

	backendAddr = "localhost:50051"
)

func callUnaryEcho(c ecpb.EchoClient, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.UnaryEcho(ctx, &ecpb.EchoRequest{Message: message})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(r.Message)
}

func makeRPCs(cc *grpc.ClientConn, n int, options ...string) {
	hwc := ecpb.NewEchoClient(cc)
	for i := 0; i < n; i++ {
		callUnaryEcho(hwc, "this is examples/name_resolving:"+strings.Join(options, "-"))
	}
}

func passThroughDemo() {
	// passthrough resolver
	passthroughConn, err := grpc.Dial(
		fmt.Sprintf("passthrough:///%s", backendAddr), // Dial to "passthrough:///localhost:50051"
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer passthroughConn.Close()

	fmt.Printf("--- calling helloworld.Greeter/SayHello to \"passthrough:///%s\"\n", backendAddr)
	makeRPCs(passthroughConn, 10)
}

// examplepassThroughDemo 使用 example resolver
func examplepassThroughDemo() {
	exampleConn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName), // Dial to "example:///resolver.example.grpc.io"
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer exampleConn.Close()

	fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s:///%s\"\n", exampleScheme, exampleServiceName)
	makeRPCs(exampleConn, 10)
}

func basicDNSDemo() {
	//如果使用 dns://192.168.1.42:443/ 则会报错 Servname not supported for ai_socktype
	target := "dns:///192.168.1.42:443"
	dnsConn, err := grpc.DialContext(context.Background(),
		target,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln("failed in dial:", err)
	}
	defer dnsConn.Close()
	fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
	makeRPCs(dnsConn, 10)
}

func consulDNSDemo() {
	// consul dns resolver
	//使用 consul 的注册中心和 dns SRV 解析功能
	if false {
		target := "consul_dns://127.0.0.1:8600/greet.service.consul" //dns://localhost:8600"
		dnsConn, err := grpc.DialContext(context.Background(),
			target,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithDisableServiceConfig(),
			grpc.WithBalancerName(roundrobin.Name),
		)
		if err != nil {
			log.Fatalln("failed in dial:", err)
		}
		defer dnsConn.Close()
		fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
		makeRPCs(dnsConn, 10)
	}
}

// customDemo
// 完全借鉴 example 的实现方法
func customDemo() {
	target := fmt.Sprintf("custom:///%s", customServiceName)
	customConn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	defer customConn.Close()
	fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
	makeRPCs(customConn, 10)
}

// etcdDemo
func etcdDemo(options ...string) {
	target := fmt.Sprintf("etcd:///%s", etcdServiceName)
	conn, err := grpc.DialContext(
		context.Background(),
		target,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
	)
	if err != nil {
		log.Fatalln("failed in dial:", err)
	}
	defer conn.Close()
	fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
	makeRPCs(conn, 10, options...)
}

func etcdRoundRobin() {
	{
		target := fmt.Sprintf("etcd:///%s", etcdServiceName)
		conn, err := grpc.DialContext(
			context.Background(),
			target,
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithBalancerName(roundrobin.Name),
		)
		if err != nil {
			log.Fatalln("failed in dial:", err)
		}
		defer conn.Close()
		fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
		makeRPCs(conn, 10, "first")
	}
	{
		target := fmt.Sprintf("etcd:///%s", etcdServiceName)
		conn, err := grpc.DialContext(
			context.Background(),
			target,
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithBalancerName(roundrobin.Name),
		)
		if err != nil {
			log.Fatalln("failed in dial:", err)
		}
		defer conn.Close()
		fmt.Printf("--- calling helloworld.Greeter/SayHello to \"%s\"\n", target)
		makeRPCs(conn, 10, "second")
	}
	fmt.Println("round robin call finished")
}

func main() {
	examplepassThroughDemo()
}

// Following is an example name resolver. It includes a
// ResolverBuilder(https://godoc.org/google.golang.org/grpc/resolver#Builder)
// and a Resolver(https://godoc.org/google.golang.org/grpc/resolver#Resolver).
//
// A ResolverBuilder is registered for a scheme (in this example, "example" is
// the scheme). When a ClientConn is created for this scheme, the
// ResolverBuilder will be picked to build a Resolver. Note that a new Resolver
// is built for each ClientConn. The Resolver will watch the updates for the
// target, and send updates to the ClientConn.

// exampleResolverBuilder is a
// ResolverBuilder(https://godoc.org/google.golang.org/grpc/resolver#Builder).
type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: {backendAddr},
		},
	}
	r.start()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

// exampleResolver is a
// Resolver(https://godoc.org/google.golang.org/grpc/resolver#Resolver).
type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}

func init() {
	// Register the example ResolverBuilder. This is usually done in a package's
	// init() function.
	resolver.Register(&exampleResolverBuilder{})
}
