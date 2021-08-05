package grpc

import (
	"eapp/component"
	app_grpc "eapp/component/rpc/grpc"
	"eapp/consts"
	"eapp/global_config"
	"eapp/pb"
	"eapp/registry"
	"eapp/registry/etcd"
	"eapp/servers"
	"eapp/utils"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
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

//接入grpc
type RpcServer struct {
	c        component.Container
	rpc      *grpc.Server
	registry registry.IRegistry
	conf     *RpcConfig
	running  string
	services map[string]string //服务的地址存放集合
	host     string
	servers  map[string]app_grpc.RpcHandler
}

func NewRpcServer(c component.Container) (servers.IServer, error) {

	conf, err := NewRpcConfig(global_config.Conf)
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
		servers:  make(map[string]app_grpc.RpcHandler),
		c:        c,
	}
	return e, nil
}

func (r *RpcServer) Start() error {
	r.c.Debug("RPC服务器正在启动...")

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

	pb.RegisterRPCServer(r.rpc, &app_grpc.RequestService{Servers: r.servers})

	if err := r.run(); err != nil {
		r.c.Errorf("服务器启动失败,err:%v", err)
		return err
	}

	r.c.Debugf("RPC服务器启动成功，addr->[ :%s ]", r.conf.RpcPort)

	return nil
}

func (r *RpcServer) run() error {

	r.running = ST_RUNNING
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", net.JoinHostPort(r.host, r.conf.RpcPort))
		if err != nil {
			ch <- fmt.Errorf("servers start error:%v", err)
		}

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
		panic("服务名为空")
		return nil
	}

	host, err := utils.GetRealIp()
	if err != nil {
		return err
	}

	r.services[r.conf.Cluster+"/"+service] = host + ":" + r.conf.RpcPort
	if _, ok := h.(app_grpc.RpcHandler); !ok {
		panic("服务类型错误")
		return nil
	}
	r.servers[service] = h.(app_grpc.RpcHandler)

	return nil
}

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
