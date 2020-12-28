package servers

import (
	"erpc/config"
	"erpc/iservers"
	"erpc/pb"
	"erpc/plugins"
	"erpc/plugins/etcd"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/naming"
	"net"
	"time"
)

func init() {
	logConf := config.LoggerConfig{}
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
	conf     *config.ServerConfig
	running  string
}

func NewErpcServer(conf *config.ServerConfig) (*ErpcServer, error) {
	rgy, err := etcd.NewEtcdRegistry(conf)
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
		conf:     conf,
	}
	return erpc, nil
}

func (s *ErpcServer) Start() error {

	//	注册服务到注册中心
	if err := s.registry.Register(s.conf.Cluster, s.conf.ServiceName, naming.Update{
		Op:       naming.Add,
		Addr:     s.conf.ServiceAddr,
		Metadata: s.conf.ServiceAddr,
	}); err != nil {
		return fmt.Errorf("注册服务到注册中心失败,err:%v", err)
	}

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
		lis, err := net.Listen("tcp", net.JoinHostPort(s.conf.RpcHost, s.conf.Port))
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

func (s *ErpcServer) RegistService(clusterName, serviceName string, h iservers.IHandler) {
	s.conf.Cluster = clusterName
	s.conf.ServiceName = serviceName
	pb.RegisterRPCServer(s.server, h)
}
