package etcd

import (
	"github.com/coreos/etcd/clientv3"
	custom_resolver "google.golang.org/grpc/examples/features/name_resolving/client/resolver"
	"google.golang.org/grpc/resolver"

	"sync"

	"fmt"

	"context"

	"log"

	"encoding/json"
)

var client *clientv3.Client
var mutex sync.Mutex

func getEtcdClient(address string) (*clientv3.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if client == nil {
		var err error
		if client, err = clientv3.New(clientv3.Config{
			Endpoints: []string{address},
		}); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func init() {
	var _ resolver.Resolver = &Resolver{}
	var _ resolver.Builder = &Builder{}
}

type Builder struct{}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if target.Authority == "" {
		target.Authority = "localhost:2379"
	}
	client, err := getEtcdClient(target.Authority)
	if err != nil {
		return nil, err
	}
	rsv := &Resolver{
		target: target,
		cc:     cc,
		client: client,
		cChan:  make(chan struct{}),
	}
	rsv.start()
	return rsv, nil
}

func (b *Builder) Scheme() string {
	return "etcd"
}

type Resolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	// opts     resolver.BuildOptions
	client *clientv3.Client
	cChan  chan struct{}
}

func (r *Resolver) start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchC := r.client.Watch(
		ctx,
		r.target.Endpoint,
		clientv3.WithPrefix(),
	)
	log.Printf("etcd watching %s\n", r.target.Endpoint)
	go func() {
	loop:
		for {
			all, err := getAllAddress(r.client, r.target.Endpoint)
			if err != nil {
				log.Fatalln("failed in get all address:", err)
			}
			r.cc.UpdateState(resolver.State{Addresses: all})
			select {
			case response := <-watchC:
				if err := response.Err(); err != nil {
					log.Fatalln(err)
				}
			case <-r.cChan:
				break loop
			}
		}

	}()
}

func getAllAddress(client *clientv3.Client, endpoint string) ([]resolver.Address, error) {
	getResponse, err := client.Get(context.TODO(), endpoint, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addresss := make([]resolver.Address, len(getResponse.Kvs))
	for i, kv := range getResponse.Kvs {
		node := custom_resolver.Node{}
		json.Unmarshal(kv.Value, &node)
		addresss[i] = resolver.Address{
			Addr: fmt.Sprintf("%v:%v", node.Host, node.Port),
		}
	}
	return addresss, nil
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	fmt.Println("add")
}

func (r *Resolver) Close() {}

func init() {
	resolver.Register(&Builder{})
}
