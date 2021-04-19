package rpc

import (
	"context"
	"eapp/balancer/random"
	"eapp/component"
	"eapp/configs"
	"eapp/consts"
	"eapp/pb"
	"eapp/registry"
	"eapp/registry/etcd"
	"eapp/servers"
	"eapp/utils"
	"encoding/json"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
	"sync"
	"time"
)

var (
	ST_RUNNING   = "running"
	ST_STOP      = "stop"
	ST_PAUSE     = "pause"
	SRV_TP_API   = "api"
	SRV_FILE_API = "file"
	SRV_TP_RPC   = "rpc"
	SRV_TP_CRON  = "cron"
	SRV_TP_MQ    = "mq"
	SRV_TP_WEB   = "web"
)

type RpcServer struct {
	c        component.Container
	rpc      *grpc.Server
	registry registry.IRegistry
	conf     *RpcConfig
	running  string
	sync.RWMutex
	connPool map[string]*grpc.ClientConn //存放rpc客户端的连接池
	services map[string]string           //服务的地址存放集合
	host     string
	servers  map[string]IHandler //存放服务的集合
}

func NewRpcServer(c component.Container) (servers.IServer, error) {

	conf, err := NewRpcConfig(configs.Conf)
	if err != nil {
		return nil, err
	}

	//构建注册中心
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: conf.RegisterTimeout,
		Username:    conf.UserName,
		Password:    conf.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("连接etcd失败,err:%v", err)
	}
	r := etcd.NewEtcdRegistry(cli, conf.TTl)

	//构建rpc，servers
	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10,
			//grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			//	UnaryServerLogInterceptor(c),
			//	grpc_ctxtags.UnaryServerInterceptor(),
			//	grpc_recovery.UnaryServerInterceptor(),
			//)),
		}),
	)

	ip, err := utils.GetRealIp()
	if err != nil {
		return nil, err
	}

	e := &RpcServer{
		host:     ip,
		rpc:      s,
		registry: r,
		conf:     conf,
		services: make(map[string]string),
		connPool: make(map[string]*grpc.ClientConn),
		servers:  make(map[string]IHandler),
		c:        c,
	}
	return e, nil
}

func (r *RpcServer) Start() error {
	r.c.Info("Rpc服务器正在启动...")

	//	注册服务到注册中心
	for key, addr := range r.services {
		if err := r.registry.Register(key, naming.Update{
			Op:       naming.Add,
			Addr:     addr,
			Metadata: addr,
		}); err != nil {
			r.c.Errorf("注册服务到注册中心失败,err:%v", err)
			return err
		}
	}

	// 注册服务到rpc服务器
	pb.RegisterRPCServer(r.rpc, &RequestService{servers: r.servers})

	if err := r.run(); err != nil {
		r.c.Errorf("服务器启动失败,err:%v", err)
		return err
	}

	r.c.Info("Rpc服务器启动成功...")

	return nil
}

func (r *RpcServer) run() error {

	r.running = ST_RUNNING
	//协程内捕获错误
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", net.JoinHostPort(r.host, r.conf.RpcPort))
		if err != nil {
			ch <- fmt.Errorf("servers start error:%v", err)
		}

		//	开启rpc服务
		if err := r.rpc.Serve(lis); err != nil {
			ch <- fmt.Errorf("rpc serve error:%v", err)
		}

	}(errChan)

	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		r.running = ST_STOP
		return err
	}
}

func (r *RpcServer) Stop() {
	if r.rpc != nil {
		r.running = ST_STOP
		r.rpc.GracefulStop() //理解为安全关闭
		time.Sleep(time.Second)
	}
	r.c.Warn("Rpc服务器已经安全关闭...")
}

//注册rpc服务 addrs: ip+port
func (r *RpcServer) RegisterService(service string, h interface{}) error {

	if service == "" {
		return fmt.Errorf("服务名为空")
	}

	host, err := utils.GetRealIp()
	if err != nil {
		return err
	}

	r.services[r.conf.Cluster+"/"+service] = host + ":" + r.conf.RpcPort
	if _, ok := h.(IHandler); !ok {
		return fmt.Errorf("服务类型错误")
	}
	r.servers[service] = h.(IHandler)

	return nil
}

//根据集群名和服务名进行调用
func (r *RpcServer) Request(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error) {

	//根据集群名和服务名获取rpc服务
	client, err := r.getService(cluster, service)
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
	i, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if len(i) == 0 {
		i = []byte("{}")
	}

	//进行调用
	resp, err := client.Request(context.Background(), &pb.RequestContext{
		Service: service,
		Input:   string(i),
		Header:  string(h),
	},
		grpc.FailFast(failFast))
	grpc.WithInsecure()
	if err != nil {
		r.c.Error("调用rpc失败,err:%v", err)
		return nil, err
	}

	return resp, nil
}

func (r *RpcServer) getService(cluster, service string) (pb.RPCClient, error) {
	//先取rpc连接,如果没有指定集群名就获取当前server的集群名
	if cluster == "" {
		cluster = r.conf.Cluster
	}
	conn, err := r.getConn("/" + cluster + "/" + service)
	if err != nil {
		return nil, err
	}
	//由连接构建rpc服务
	return pb.NewRPCClient(conn), nil
}

//serviceName: "/cluster/service"
func (r *RpcServer) getConn(serviceName string) (*grpc.ClientConn, error) {
	//根据service来获取现有的连接
	r.RLock()
	if conn, ok := r.connPool[serviceName]; ok {
		r.RUnlock()
		return conn, nil
	}
	r.RUnlock()
	r.Lock()
	defer r.Unlock()
	ecli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   r.conf.Endpoints,
		DialTimeout: r.conf.RegisterTimeout,
	})
	if err != nil {
		return nil, err
	}

	resolver := etcd.NewEtcdRegistry(ecli, r.conf.TTl)

	dialOpts := []grpc.DialOption{
		grpc.WithTimeout(r.conf.RpcTimeout),
		grpc.WithBalancer(r.getBalancer(r.conf.BalanceMod, resolver)),
		grpc.WithInsecure(),
		grpc.WithBlock(), //表示所有的rpc调用顺序进行
	}
	//根据负载均衡获取连接(负载均衡器去同一个服务名前缀下的所有节点筛选)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	conn, err := grpc.Dial(serviceName, dialOpts...)
	if err != nil {
		return nil, err
	}

	r.connPool[serviceName] = conn
	return conn, nil
}

func (r *RpcServer) getBalancer(mod int8, lb naming.Resolver) grpc.Balancer {
	switch mod {
	case RoundRobin:
		return grpc.RoundRobin(lb)
	case Random:
		return random.Random(lb)
	default:
		return grpc.RoundRobin(lb)
	}
}

//适配器模式构建rpc服务器
type rpcServerAdapter struct{}

func (*rpcServerAdapter) Resolve(c component.Container) servers.IServer {

	rpsServer, err := NewRpcServer(c)
	if err != nil {
		panic(err)
	}
	return rpsServer
}

func init() {
	servers.Register(consts.RpcServer, &rpcServerAdapter{})
}
