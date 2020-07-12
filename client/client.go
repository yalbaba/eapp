package client

import (
	"erpc/plugins/retcd"
	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type ErpcClient struct {
	sync.RWMutex
	ConnPool     map[string]*grpc.ClientConn
	DialOpts     []grpc.DialOption
	RegisterAddr string
}

func NewErpcClient(opts []grpc.DialOption, registerAddr string) *ErpcClient {
	return &ErpcClient{
		ConnPool:     make(map[string]*grpc.ClientConn),
		DialOpts:     opts,
		RegisterAddr: registerAddr,
	}
}

func (c *ErpcClient) GetConn(serviceName, serverAddr string, registerAddrs []string) (*grpc.ClientConn, error) {
	//根据service来获取现有的连接
	c.RLock()
	if conn, ok := c.ConnPool[serviceName]; ok {
		c.RUnlock()
		return conn, nil
	}

	c.Lock()
	defer c.Unlock()

	//	resolver
	ecli, err := etcd.New(etcd.Config{
		Endpoints:   registerAddrs,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver := retcd.NewEtcdRegistry(ecli)
	//	负载均衡暂时用grpc的轮询
	balancer := grpc.RoundRobin(resolver)

	dialOpts := []grpc.DialOption{grpc.WithTimeout(time.Second * 3), grpc.WithBalancer(balancer), grpc.WithInsecure(), grpc.WithBlock()}

	conn, err := grpc.Dial(serverAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	c.ConnPool[serviceName] = conn
	return conn, nil
}
