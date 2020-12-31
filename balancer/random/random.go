package random

import (
	"context"
	"erpc/balancer"
	"erpc/enum"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
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

func Random(r naming.Resolver) grpc.Balancer {
	return &RandomBalancer{r: r}
}

//开启负载均衡器
//target: /cluster/service
func (r *RandomBalancer) Start(target string, config grpc.BalancerConfig) error {
	r.Lock()
	if r.done {
		r.Unlock()
		return grpc.ErrClientConnClosing
	}
	if r.r == nil {
		r.Unlock()
		return fmt.Errorf("register is closed")
	}
	watcher, err := r.r.Resolve(target)
	if err != nil {
		return err
	}
	r.w = watcher
	r.addrCh = make(chan []grpc.Address)
	r.Unlock()

	//开启监听服务地址是否发生改变
	go func() {
		if err := r.watchAddrUpdates(); err != nil {

			return
		}
	}()
	return nil
}

func (r *RandomBalancer) watchAddrUpdates() error {
	updates, err := r.w.Next()
	if err != nil {
		return err
	}
	r.Lock()
	defer r.Unlock()
	for _, update := range updates {
		//得到新的操作的地址对象
		addr := grpc.Address{
			Addr:     update.Addr,
			Metadata: update.Metadata,
		}

		switch update.Op {
		case naming.Add:
			//新增服务地址
			//	将新的服务地址加到注册中心里
			var exist bool
			for _, v := range r.addrs {
				if addr == v.Addr {
					exist = true
					break
				}
			}
			if exist {
				//当前地址已存在，则继续判断下一个更新的地址
				continue
			}
			//	不存在则加到原地址集合里
			r.addrs = append(r.addrs, &balancer.AddrInfo{
				Addr: addr,
			})
		case enum.Modify:
			if !r.weight {
				continue
			}
			//	更新权重
			for _, v := range r.addrs {
				if addr.Addr == v.Addr.Addr {
					//v.Weight = balancer.GetWeightByMetadata(addr.Metadata)
					break
				}
			}
		case naming.Delete:
			//	删除操作
			for i, v := range r.addrs {
				if addr.Addr == v.Addr.Addr {
					//相当于除去第i个位置的数据
					copy(r.addrs[i:], r.addrs[i+1:])
					r.addrs = r.addrs[:len(r.addrs)-1]
					break
				}
			}
		default:
			grpclog.Errorf("Unknown update.Op :%d", update.Op)
		}
	}
	//	因为可能操作了地址集合，所以进行通知
	open := make([]grpc.Address, len(r.addrs))
	for _, v := range r.addrs {
		open = append(open, v.Addr)
	}
	if r.done {
		return grpc.ErrClientConnClosing
	}
	select {
	case <-r.addrCh:
	default:
	}

	r.addrCh <- open
	return nil
}

func (r *RandomBalancer) Up(addr grpc.Address) (down func(error)) {
	return nil
}

func (r *RandomBalancer) Get(ctx context.Context, opts grpc.BalancerGetOptions) (addr grpc.Address, put func(), err error) {
	return grpc.Address{}, nil, nil
}

func (r *RandomBalancer) Notify() <-chan []grpc.Address {
	return nil
}
func (r *RandomBalancer) Close() error {
	return nil
}
