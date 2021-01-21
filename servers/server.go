package servers

import (
	"context"
	"encoding/json"
	"erpc/balancer/random"
	"erpc/iservers"
	"erpc/pb"
	"erpc/plugins"
	"erpc/plugins/etcd"
	"erpc/utils"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
	"sync"
	"time"
)

//func init() {
//	logConf := global.LoggerConfig{}
//	grpclog.SetLoggerV2(logConf.NewLogger())
//}

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

type ErpcServer struct {
	server   *grpc.Server
	registry plugins.IRegistry
	conf     *RpcConfig
	running  string
	sync.RWMutex
	connPool map[string]*grpc.ClientConn //存放rpc客户端的连接池
	services map[string]string           //服务的地址存放集合
	host     string
	servers  map[string]iservers.Handler //存放服务的集合
}

func NewErpcServer(conf *RpcConfig) (*ErpcServer, error) {

	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   conf.RegisterAddrs,
		DialTimeout: conf.RegisterTimeOut,
		Username:    conf.UserName,
		Password:    conf.Pass,
	})
	if err != nil {
		return nil, fmt.Errorf("etcd cli init error:%v", err)
	}
	rgy, err := etcd.NewEtcdRegistry(cli)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		//grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		//	UnaryServerLogInterceptor(c),
		//	grpc_ctxtags.UnaryServerInterceptor(),
		//	grpc_recovery.UnaryServerInterceptor(),
		//)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10, //todo
		}),
	)

	iplocal, err := utils.GetRealIp()
	if err != nil {
		return nil, err
	}

	erpc := &ErpcServer{
		host:     iplocal,
		server:   server,
		registry: rgy,
		conf:     conf,
		services: make(map[string]string),
		connPool: make(map[string]*grpc.ClientConn),
		servers:  make(map[string]iservers.Handler),
	}
	return erpc, nil
}

func (s *ErpcServer) Start() error {

	//	注册服务到注册中心
	fmt.Println("s.services:::::::", s.services)
	for key, addr := range s.services {
		if err := s.registry.Register(key, naming.Update{
			Op:       naming.Add,
			Addr:     addr,
			Metadata: addr,
		}); err != nil {
			return fmt.Errorf("注册服务到注册中心失败,err:%v", err)
		}
	}

	// 注册服务到rpc服务器
	pb.RegisterRPCServer(s.server, &RequestService{servers: s.servers})

	if err := s.Run(); err != nil {
		return fmt.Errorf("开启rpc服务失败,err:%v", err)
	}

	return nil
}

func (s *ErpcServer) Run() error {

	s.running = ST_RUNNING
	//协程内捕获错误
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", net.JoinHostPort(s.host, s.conf.RpcPort))
		if err != nil {
			ch <- fmt.Errorf("servers start error:%v", err)
		}

		//	开启rpc服务
		if err := s.server.Serve(lis); err != nil {
			ch <- fmt.Errorf("rpc serve error:%v", err)
		}

	}(errChan)

	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		s.running = ST_STOP
		return err
	}
}

func (s *ErpcServer) Stop() {
	if s.server != nil {
		s.running = ST_STOP
		s.server.GracefulStop() //理解为安全关闭
		time.Sleep(time.Second)
	}
}

//注册rpc服务 addrs: ip+port
func (s *ErpcServer) RegistService(serviceName string, h iservers.Handler) error {

	if serviceName == "" {
		return fmt.Errorf("服务名为空")
	}

	iplocal, err := utils.GetRealIp()
	if err != nil {
		return err
	}
	s.services[s.conf.Cluster+"/"+serviceName] = iplocal + ":" + s.conf.RpcPort

	s.servers[serviceName] = h

	return nil
}

//根据集群名和服务名进行调用
func (s *ErpcServer) Rpc(serviceName string, input map[string]interface{}) (interface{}, error) {

	if _, ok := s.servers[serviceName]; !ok {
		return nil, fmt.Errorf("该服务未注册!")
	}

	//根据集群名和服务名获取rpc服务
	client, err := s.getService(serviceName)
	if err != nil {
		return nil, err
	}

	//进行调用
	b, _ := json.Marshal(input)
	return client.Request(context.Background(), &pb.RequestContext{
		Service: serviceName,
		Input:   string(b),
	})
}

func (s *ErpcServer) getService(serviceName string) (pb.RPCClient, error) {
	//先取rpc连接
	conn, err := s.getConn("/" + s.conf.Cluster + "/" + serviceName)
	if err != nil {
		return nil, err
	}
	//由连接构建rpc服务
	return pb.NewRPCClient(conn), nil
}

//serviceName: "/"+ cluster + "/" + service
func (e *ErpcServer) getConn(serviceName string) (*grpc.ClientConn, error) {
	//根据service来获取现有的连接
	e.RLock()
	if conn, ok := e.connPool[serviceName]; ok {
		e.RUnlock()
		return conn, nil
	}
	e.RUnlock()
	e.Lock()
	defer e.Unlock()
	ecli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   e.conf.RegisterAddrs,
		DialTimeout: e.conf.RegisterTimeOut,
	})
	if err != nil {
		return nil, err
	}

	resolver, err := etcd.NewEtcdRegistry(ecli)
	if err != nil {
		return nil, err
	}

	fmt.Println("e.conf.BalancerMod:::", e.conf.BalancerMod)
	dialOpts := []grpc.DialOption{
		grpc.WithTimeout(e.conf.RpcTimeOut),
		grpc.WithBalancer(e.getBalancer(e.conf.BalancerMod, resolver)),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}
	//根据负载均衡获取连接(负载均衡器去同一个服务名前缀下的所有节点筛选)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	conn, err := grpc.Dial(serviceName, dialOpts...)
	if err != nil {
		return nil, err
	}

	e.connPool[serviceName] = conn
	return conn, nil
}

func (e *ErpcServer) getBalancer(mod int8, r naming.Resolver) grpc.Balancer {
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
