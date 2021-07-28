package grpc

import (
	"context"
	"eapp/balancer/random"
	"eapp/component/rpc"
	"eapp/global_config"
	"eapp/pb"
	"eapp/registry/etcd"
	"encoding/json"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"sync"
	"time"
)

type Invoker struct {
	conf *global_config.Config
	sync.RWMutex
	connPool map[string]*grpc.ClientConn //存放rpc客户端的连接池
}

func NewRpcInvoker(conf *global_config.Config) rpc.RpcInvoker {
	return &Invoker{
		conf:     conf,
		connPool: make(map[string]*grpc.ClientConn),
	}
}

//根据集群名和服务名进行调用
func (i *Invoker) Request(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error) {

	//根据集群名和服务名获取rpc服务
	client, err := i.getService(cluster, service)
	if err != nil {
		return nil, err
	}

	h, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	if len(h) == 0 {
		h = []byte("{}")
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		bytes = []byte("{}")
	}

	//进行调用
	resp, err := client.Request(context.Background(), &pb.RequestContext{
		Service: service,
		Input:   string(bytes),
		Header:  string(h),
	},
		grpc.FailFast(failFast))
	grpc.WithInsecure()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (i *Invoker) getService(cluster, service string) (pb.RPCClient, error) {
	//先取rpc连接,如果没有指定集群名就获取当前server的集群名
	if cluster == "" {
		return nil, fmt.Errorf("集群名不能为空")
	}
	conn, err := i.getConn("/" + cluster + "/" + service)
	if err != nil {
		return nil, err
	}
	//由连接构建rpc服务
	return pb.NewRPCClient(conn), nil
}

//serviceName: "/cluster/service"
func (i *Invoker) getConn(serviceName string) (*grpc.ClientConn, error) {
	//根据service来获取现有的连接
	i.RLock()
	if conn, ok := i.connPool[serviceName]; ok {
		i.RUnlock()
		return conn, nil
	}
	i.RUnlock()
	i.Lock()
	defer i.Unlock()
	ecli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   i.conf.Registry.EndPoints,
		DialTimeout: time.Duration(i.conf.Registry.RegisterTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	resolver := etcd.NewEtcdRegistry(ecli, i.conf.Registry.TTl)

	dialOpts := []grpc.DialOption{
		grpc.WithBalancer(i.getBalancer(i.conf.GrpcService.BalanceMod, resolver)),
		grpc.WithInsecure(),
		grpc.WithBlock(), //表示所有的rpc调用顺序进行
	}
	//根据负载均衡获取连接(负载均衡器去同一个服务名前缀下的所有节点筛选)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(i.conf.GrpcService.RpcTimeout)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, serviceName, dialOpts...)
	if err != nil {
		return nil, err
	}

	i.connPool[serviceName] = conn
	return conn, nil
}

const (
	RoundRobin = iota + 10
	Random
)

func (i *Invoker) getBalancer(mod int8, lb naming.Resolver) grpc.Balancer {
	switch mod {
	case RoundRobin:
		return grpc.RoundRobin(lb)
	case Random:
		return random.Random(lb)
	default:
		return grpc.RoundRobin(lb)
	}
}

func (i *Invoker) Close() {
	i.RLock()
	defer i.RUnlock()
	for _, conn := range i.connPool {
		conn.Close()
	}
}

//统一的rpc请求入口
type RequestService struct {
	Servers map[string]RpcHandler
}

func (r *RequestService) Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error) {
	if _, ok := r.Servers[in.Service]; !ok {
		return &pb.ResponseContext{
			Status: 500,
			Result: "服务未注册",
		}, fmt.Errorf("服务未注册")
	}

	header := make(map[string]string)
	if err := json.Unmarshal([]byte(in.Header), &header); err != nil {
		return nil, err
	}

	input := make(map[string]interface{})
	if err := json.Unmarshal([]byte(in.Input), &input); err != nil {
		return nil, err
	}
	resp, err := r.Servers[in.Service].Handle(ctx, header, input)
	if err != nil {
		return &pb.ResponseContext{
			Status: 500,
			Result: resp.(string),
		}, err
	}

	return &pb.ResponseContext{
		Status: 500,
		Result: resp.(string),
	}, nil
}
