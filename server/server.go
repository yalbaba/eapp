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

type ErpcServer struct {
	server   grpc.Server
	registry plugins.Registry
	config   *ServerConfig
}

func NewErpcServer(conf *ServerConfig, server grpc.Server) (*ErpcServer, error) {
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
	//开tcp监听
	lis, err := net.Listen("tcp", net.JoinHostPort(s.config.RpcHost, s.config.Port))
	if err != nil {
		return fmt.Errorf("server start error:%v", err)
	}
	//	注册服务到注册中心
	s.registry.Register(context.TODO(), s.config.Cluster, s.config.ServiceName, naming.Update{
		Op:       naming.Add,
		Addr:     s.config.ServiceAddr,
		Metadata: s.config.ServiceAddr,
	})
	//	开启rpc服务
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("rpc serve error:%v", err)
	}
	return nil
}

func (s *ErpcServer) stop() {
	s.server.Stop()
}
