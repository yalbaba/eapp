package random

import (
	"context"
	"erpc/balancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"math/rand"
	"sync"
)

type RandomBalancer struct {
	sync.Mutex
	rand   *rand.Rand
	r      naming.Resolver
	w      naming.Watcher
	addrs  []*balancer.AddrInfo // 客户端可能连接的所有地址
	addrCh chan []grpc.Address  // 通知gRPC内部客户端应连接到的地址列表的通道。
	waitCh chan struct{}        // 当没有可用的连接地址时要阻止的通道
	done   bool                 // 是否关闭负载均衡器关闭
	weight bool                 // 是否按照权重做负载均衡
}

func (r *RandomBalancer) Start(target string, config grpc.BalancerConfig) error {
	return nil
}

func (r *RandomBalancer) Up(addr grpc.Address) (down func(error)) {
	return nil
}

func (r *RandomBalancer) Get(ctx context.Context, opts grpc.BalancerGetOptions) (addr grpc.Address, put func(), err error) {
	return nil, nil, nil
}

func (r *RandomBalancer) Notify() <-chan []grpc.Address {
	return nil
}
func (r *RandomBalancer) Close() error {
	return nil
}
