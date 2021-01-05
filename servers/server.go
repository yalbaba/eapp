package servers

import (
	"erpc/global"
	"erpc/iservers"
	"erpc/pb"
	"erpc/plugins"
	"erpc/plugins/etcd"
	"erpc/utils"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
	"time"
)

func init() {
	logConf := global.LoggerConfig{}
	grpclog.SetLoggerV2(logConf.NewLogger())
}

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
}

func NewErpcServer(conf *global.ServerConfig) (*ErpcServer, error) {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   conf.RegisterAddrs,
		DialTimeout: 5 * time.Second,
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
			Time: 10,
		}),
	)

	erpc := &ErpcServer{
		server:   server,
		registry: rgy,
		conf:     NewRpcConfig(conf),
	}
	return erpc, nil
}

func (s *ErpcServer) Start() error {
	// 参数校验
	if err := s.check(); err != nil {
		return err
	}

	//	注册服务到注册中心
	for key, addr := range s.conf.Services {
		if err := s.registry.Register(key, naming.Update{
			Op:       naming.Add,
			Addr:     addr,
			Metadata: addr,
		}); err != nil {
			return fmt.Errorf("注册服务到注册中心失败,err:%v", err)
		}
	}

	if err := s.Run(); err != nil {
		return fmt.Errorf("开启rpc服务失败,err:%v", err)
	}

	return nil
}

func (s *ErpcServer) check() error {

	if s.conf.Cluster == "" {
		return fmt.Errorf("集群名不能为空")
	}
	if len(s.conf.Services) == 0 {
		return fmt.Errorf("服务集合为空")
	}

	return nil
}

func (s *ErpcServer) Run() error {

	s.running = ST_RUNNING
	//协程内捕获错误
	errChan := make(chan error, 1)
	go func(ch chan error) {
		lis, err := net.Listen("tcp", net.JoinHostPort(s.conf.RpcAddr, s.conf.RpcPort))
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
func (s *ErpcServer) RegistService(cluster, serviceName string, h iservers.Handler, input map[string]interface{}) error {

	if cluster != "" {
		s.conf.Cluster = cluster
	}

	if serviceName == "" {
		return fmt.Errorf("服务名为空")
	}

	s.conf.Services["/"+cluster+"/"+serviceName] = utils.GetRealIp()
	pb.RegisterRPCServer(s.server, &RequestService{
		input:  input,
		handle: h,
	})

	return nil
}
