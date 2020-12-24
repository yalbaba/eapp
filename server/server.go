package server

import (
	"context"
	"erpc/config"
	"erpc/plugins"
	"erpc/plugins/etcd"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
	"net"
	"time"
)

func init() {
	logConf := config.LoggerConfig{}
	grpclog.SetLoggerV2(logConf.NewLogger())
}

type ServerConfig struct {
	Cluster       string
	ServiceName   string
	ServiceAddr   string
	RegisterAddrs []string
	RpcHost       string
	Port          string
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
	registry plugins.Registry
	config   *ServerConfig
	running  string
}

func NewErpcServer(conf *ServerConfig, server *grpc.Server) (*ErpcServer, error) {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   conf.RegisterAddrs,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("etcd cli init error:%v", err)
	}
	registry := etcd.NewEtcdRegistry(cli)
	erpc := &ErpcServer{
		server:   server,
		registry: registry,
		config:   conf,
	}
	return erpc, nil
}

func (s *ErpcServer) Start() error {

	//	注册服务到注册中心
	if err := s.registry.Register(context.TODO(), s.config.Cluster, s.config.ServiceName, naming.Update{
		Op:       naming.Add,
		Addr:     s.config.ServiceAddr,
		Metadata: s.config.ServiceAddr,
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
		lis, err := net.Listen("tcp", net.JoinHostPort(s.config.RpcHost, s.config.Port))
		if err != nil {
			ch <- fmt.Errorf("server start error:%v", err)
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

func (s *ErpcServer) stop() {
	if s.server != nil {
		s.running = ST_STOP
		s.server.GracefulStop() //理解为安全关闭
		time.Sleep(time.Second)
	}
}
