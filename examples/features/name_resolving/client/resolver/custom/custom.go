package custom

import (
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&Builder{})
}

// Builder 自定义的 Builder
type Builder struct{}

// Build 构建 Resolver
func (c *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	rsv := &Resolver{
		target: target,
		cc:     cc,
		opts:   opts,
		addrStore: map[string][]string{
			"resolver.custom": []string{
				"192.168.1.42:50051",
			},
		},
	}
	rsv.start()
	return rsv, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (c *Builder) Scheme() string {
	return "custom"
}

type Resolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	opts      resolver.BuildOptions
	addrStore map[string][]string
}

func (r *Resolver) start() {
	addrs := r.addrStore[r.target.Endpoint]
	newAddr := make([]resolver.Address, len(addrs))
	for i, addr := range addrs {
		newAddr[i] = resolver.Address{Addr: addr}
	}
	r.cc.UpdateState(resolver.State{
		Addresses: newAddr,
	})
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {}
