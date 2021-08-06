package random

import (
	"context"
	"eapp/balancer"
	"eapp/utils"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"
	"math/rand"
	"sync"
)

//带权重的随机负载均衡
type random struct {
	sync.Mutex
	rand   *rand.Rand
	r      naming.Resolver
	w      naming.Watcher
	addrs  []balancer.AddrInfo // 客户端可能连接的所有地址
	addrCh chan []grpc.Address // 通知gRPC内部客户端应连接到的地址列表的通道。
	waitCh chan struct{}       // 当没有可用的连接地址时要阻止的通道
	done   bool                // 是否关闭负载均衡器关闭
	weight bool                // 是否按照权重做负载均衡
	next   int                 // index of the next address to return for Get()
}

func Random(r naming.Resolver) grpc.Balancer {
	return &random{r: r}
}

//开启负载均衡器  target: /cluster/service
func (r *random) Start(target string, config grpc.BalancerConfig) error {

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

//监控服务节点变化
func (r *random) watchAddrUpdates() error {
	updates, err := r.w.Next()
	if err != nil {
		grpclog.Warningf("grpc: the naming watcher stops working due to %v.", err)
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
			//	将新的服务地址加到注册中心里
			var exist bool
			for _, v := range r.addrs {
				if addr == v.Addr {
					exist = true
					grpclog.Infoln("grpc: The name resolver wanted to add an existing address: ", addr)
					break
				}
			}
			if exist {
				//当前地址已存在，则继续判断下一个更新的地址
				continue
			}
			//	不存在则加到原地址集合里
			r.addrs = append(r.addrs, balancer.AddrInfo{Addr: addr})
		//case _const.Modify:
		//	if !r.weight {
		//		continue
		//	}
		//	//	更新权重
		//	for _, v := range r.addrs {
		//		if addr.Addr == v.Addr.Addr {
		//			//v.Weight = balancer.GetWeightByMetadata(addr.Metadata)
		//			break
		//		}
		//	}
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
	//	地址集合，所以进行通知
	open := make([]grpc.Address, len(r.addrs))
	for i, v := range r.addrs {
		open[i] = v.Addr
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

/*
Up方法是gRPC内部负载均衡的watcher调用的，该watcher会读全局的连接状态队列，
改变RoundRobin维护的连接列表的里连接的状态（会有单独的goroutine向目标服务发起连接尝试
，尝试成功后才会把连接对象的连接状态改为connected），
如果是已连接状态的连接 ，会调用Up方法来改变addrs地址数组中该地址的状态为已连接。
*/
func (r *random) Up(addr grpc.Address) func(error) {
	//判断列表里保存的地址是否是连接状态，如果已经是，则返回，不是就更新为连接状态
	r.Lock()
	defer r.Unlock()

	var cnt int //表示当前总共已经连接的地址的数量
	for _, a := range r.addrs {
		if a.Addr == addr {
			if a.Connected {
				return nil
			}
			a.Connected = true
		}
		if a.Connected {
			cnt++
		}
	}

	// 当有一个可用地址时，之前可能是0个，可能要很多client阻塞在获取连接地址上，这里通知所有的client有可用连接啦。
	// 为什么只等于1时通知？因为可用地址数量>1时，client是不会阻塞的。
	if cnt == 1 && r.waitCh != nil {
		close(r.waitCh)
		r.waitCh = nil
	}

	//返回禁用该地址的方法
	return func(err error) {
		r.down(addr, err)
	}
}

/*
客户端在调用gRPC具体Method的Invoke方法里，会去RoundRobin的连接池addrs里获取连接，
如果addrs为空，或者addrs里的地址都不可用，Get()方法会返回错误。
但是如果设置了failfast = false，Get()方法会阻塞在waitCh这个通道上，直至Up方法给到通知，然后轮询调度可用的地址。
*/
func (r *random) Get(ctx context.Context, opts grpc.BalancerGetOptions) (addr grpc.Address, put func(), err error) {
	var ch chan struct{}
	r.Lock()
	if r.done {
		err = grpc.ErrClientConnClosing
		r.Unlock()
		return
	}

	//开始进行普通随机算法获取 todo 下个版本迭代加权随机算法
	if len(r.addrs) > 0 {
		searched := make(map[int]bool)
		for {
			a := r.addrs[r.next]                                         //上一轮索引位置的地址
			index := int(utils.Rand.RandStart(0, int64(len(r.addrs)-1))) //随机索引
			if a.Connected {
				addr = a.Addr
				r.next = index
				r.Unlock()
				return
			}
			if len(searched) == len(r.addrs) {
				break
			}
			searched[r.next] = true
			r.next = index
		}
	}

	/*
		分为阻塞模式和非阻塞模式的处理
	*/

	if !opts.BlockingWait {
		//如果是非阻塞模式，如果没有可用地址，那么报错
		if len(r.addrs) == 0 {
			err = status.Errorf(codes.Unavailable, "没有可用地址")
			r.Unlock()
			return
		}
		// Returns the next addr on rr.addrs for failfast RPCs.
		addr = r.addrs[r.next].Addr
		r.next = int(utils.Rand.RandStart(0, int64(len(r.addrs)-1)))
		r.Unlock()
		return
	}

	//如果是阻塞模式，那么需要阻塞在waitCh上，直到Up方法给通知
	if r.waitCh == nil {
		ch = make(chan struct{})
		r.waitCh = ch
	} else {
		ch = r.waitCh
	}
	r.Unlock()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-ch:
			//重复之前的随机算法操作
			r.Lock()
			if r.done {
				err = grpc.ErrClientConnClosing
				r.Unlock()
				return
			}

			//开始进行普通随机算法获取 todo 下个版本迭代加权随机算法
			if len(r.addrs) > 0 {
				index := r.next
				for {
					searched := make(map[int]bool)
					a := r.addrs[index]
					index = int(utils.Rand.RandStart(0, int64(len(r.addrs)-1)))
					if a.Connected {
						addr = a.Addr
						r.Unlock()
						return
					}
					if len(searched) == len(r.addrs) && index == r.next {
						break
					}
				}
			}

			// 地址有可能被down掉了，所以重新构建等待通道，下一次还可能回去这个链接
			if r.waitCh == nil {
				ch = make(chan struct{})
				r.waitCh = ch
			} else {
				ch = r.waitCh
			}
			r.Unlock()
		}

	}
}

func (r *random) Notify() <-chan []grpc.Address {
	return r.addrCh
}

func (r *random) Close() error {
	r.Lock()
	defer r.Unlock()

	if r.done {
		return errors.New("erpc: balancer is closed")
	}
	r.done = true
	if r.w != nil {
		r.w.Close()
	}
	if r.waitCh != nil {
		close(r.waitCh)
	}
	if r.addrCh != nil {
		close(r.addrCh)
	}
	return nil
}

// 上线服务的错误回调函数
func (r *random) down(addr grpc.Address, err error) {
	r.Lock()
	defer r.Unlock()

	for _, a := range r.addrs {
		if addr == a.Addr {
			a.Connected = false
		}
	}
}
