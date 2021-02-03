package servers

import (
	"context"
	"encoding/json"
	"erpc/balancer/random"
	"erpc/logger"
	"erpc/pb"
	"erpc/registry"
	"erpc/registry/etcd"
	"erpc/utils"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
	"os"
	"os/signal"
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

type ErpcServer struct {
	server   *grpc.Server
	registry registry.IRegistry
	conf     *RpcConfig
	running  string
	sync.RWMutex
	connPool map[string]*grpc.ClientConn //存放rpc客户端的连接池
	services map[string]string           //服务的地址存放集合
	host     string
	servers  map[string]Handler //存放服务的集合
	log      logger.ILogger
}

func NewErpcServer(conf *RpcConfig) (*ErpcServer, error) {

	//构建注册中心
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   conf.RegisterAddrs,
		DialTimeout: conf.RegisterTimeOut,
		Username:    conf.UserName,
		Password:    conf.Password,
	})
	if err != nil {
		return nil, err
	}
	r, err := etcd.NewEtcdRegistry(cli, conf.TTl)
	if err != nil {
		return nil, err
	}

	//构建rpc，server
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
		registry: r,
		conf:     conf,
		services: make(map[string]string),
		connPool: make(map[string]*grpc.ClientConn),
		servers:  make(map[string]Handler),
		log:      conf.log,
	}
	return erpc, nil
}

func (s *ErpcServer) Start() error {
	s.log.Info("服务器正在启动...")

	//	注册服务到注册中心
	for key, addr := range s.services {
		if err := s.registry.Register(key, naming.Update{
			Op:       naming.Add,
			Addr:     addr,
			Metadata: addr,
		}); err != nil {
			s.log.Errorf("注册服务到注册中心失败,err:%v", err)
			return err
		}
	}

	// 注册服务到rpc服务器
	pb.RegisterRPCServer(s.server, &RequestService{servers: s.servers})

	if err := s.run(); err != nil {
		s.log.Errorf("服务器启动失败,err:%v", err)
		return err
	}

	//监听关闭信号
	signalCh := make(chan os.Signal, 1)
	closeCh := make(chan bool)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for _ = range signalCh { //遍历捕捉到的Ctrl+C信号
			s.log.Warn("正在关闭服务器...")
			s.Stop()
			closeCh <- true
		}
	}()
	<-closeCh //阻塞进程

	return nil
}

func (s *ErpcServer) run() error {

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
	s.log.Warn("服务器已经安全关闭...")
}

//注册rpc服务 addrs: ip+port
func (s *ErpcServer) RegistService(serviceName string, h Handler) error {

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

	//根据集群名和服务名获取rpc服务
	client, err := s.getService(serviceName)
	if err != nil {
		return nil, err
	}

	//进行调用
	b, _ := json.Marshal(input)
	resp, err := client.Request(context.Background(), &pb.RequestContext{
		Service: serviceName,
		Input:   string(b),
	})
	if err != nil {
		s.log.Error("调用rpc失败,err:%v", err)
		return nil, err
	}

	return resp, nil
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

	resolver, err := etcd.NewEtcdRegistry(ecli, e.conf.TTl)
	if err != nil {
		return nil, err
	}

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
