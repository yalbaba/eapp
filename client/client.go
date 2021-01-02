package client

import (
	"erpc/balancer/random"
	"erpc/plugins/etcd"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"sync"
	"time"
)

type ErpcClient struct {
	sync.RWMutex
	ConnPool     map[string]*grpc.ClientConn
	RegisterAddr string
	conf         *ClientConfig
}

func NewErpcClient(registerAddr string, opts ...option) *ErpcClient {
	options := &ClientOptions{}
	for _, o := range opts {
		o(options)
	}
	return &ErpcClient{
		ConnPool:     make(map[string]*grpc.ClientConn),
		RegisterAddr: registerAddr,
		conf: &ClientConfig{
			RpcTimeOut:      options.RpcTimeOut,
			BalancerMod:     options.BalancerMod,
			RegisterTimeOut: options.RegisterTimeOut,
		},
	}
}

func (c *ErpcClient) GetBalancer(mod int8, r naming.Resolver) grpc.Balancer {
	//	根据配置来决定使用哪种负载均衡
	switch mod {
	case RoundRobin:
		return grpc.RoundRobin(r)
	case Random:
		return random.Random(r)
	default:
		//默认轮询
		return grpc.RoundRobin(r)
	}
}

//serviceName: "/"+ cluster + "/" + service
func (c *ErpcClient) GetConn(serviceName string, registerAddrs []string) (*grpc.ClientConn, error) {

	//根据service来获取现有的连接
	c.RLock()
	if conn, ok := c.ConnPool[serviceName]; ok {
		c.RUnlock()
		return conn, nil
	}

	c.Lock()
	defer c.Unlock()

	ecli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   registerAddrs,
		DialTimeout: time.Duration(c.conf.RegisterTimeOut.Second()),
	})
	if err != nil {
		return nil, err
	}

	resolver, err := etcd.NewEtcdRegistry(ecli)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTimeout(time.Duration(c.conf.RpcTimeOut.Second())),
		grpc.WithBalancer(c.GetBalancer(c.conf.BalancerMod, resolver)),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	//根据负载均衡获取连接(负载均衡器去同一个服务名前缀下的所有节点筛选)
	conn, err := grpc.Dial(serviceName, dialOpts...)
	if err != nil {
		return nil, err
	}

	c.ConnPool[serviceName] = conn
	return conn, nil
}
