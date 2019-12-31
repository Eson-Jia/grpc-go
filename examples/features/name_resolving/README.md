# Name resolving

This examples shows how `ClientConn` can pick different name resolvers.

## What is a name resolver

A name resolver can be seen as a `map[service-name][]backend-ip`. It takes a
service name, and returns a list of IPs of the backends. A common used name
resolver is DNS.

In this example, a resolver is created to resolve `resolver.example.grpc.io` to
`localhost:50051`.

## Try it

```
go run server/main.go
```

```
go run client/main.go
```

## Explanation

The echo server is serving on ":50051". Two clients are created, one is dialing
to `passthrough:///localhost:50051`, while the other is dialing to
`example:///resolver.example.grpc.io`. Both of them can connect the server.

Name resolver is picked based on the `scheme` in the target string. See
https://github.com/grpc/grpc/blob/master/doc/naming.md for the target syntax.

The first client picks the `passthrough` resolver, which takes the input, and
use it as the backend addresses.

The second is connecting to service name `resolver.example.grpc.io`. Without a
proper name resolver, this would fail. In the example it picks the `example`
resolver that we installed. The `example` resolver can handle
`resolver.example.grpc.io` correctly by returning the backend address. So even
though the backend IP is not set when ClientConn is created, the connection will
be created to the correct backend.

## debug

这次`consul`还是报警告,应该还是未能解析出来,能正常`RPC`是因为使用了默认的`443`端口，而恰巧 server 监听了默认`443`端口。

```log
[DEBUG] consul: Skipping self join check for "ubuntu-virtual-machine" since the cluster is too small
[DEBUG] dns: request for name greet.service.consul. type AAAA class IN (took 193.834µs) from client 127.0.0.1:44734 (udp)
[DEBUG] dns: request for name greet.service.consul. type A class IN (took 128.377µs) from client 127.0.0.1:33194 (udp)
[DEBUG] dns: request for name _grpc_config.greet.service.consul. type TXT class IN (took 120.114µs) from client 127.0.0.1:50956 (udp)
[WARN] dns: QName invalid: _grpc_config.greet.service.consul.lan
[DEBUG] dns: request for name _grpc_config.greet.service.consul.lan. type TXT class IN (took 67.833µs) from client 127.0.0.1:35403 (udp)
```
