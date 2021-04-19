package rpc

import (
	"eapp/configs"
	"fmt"
	"time"
)

const (
	RoundRobin = iota + 10
	Random
)

//rpc服务的配置
type RpcConfig struct {
	Cluster         string
	RpcPort         string
	RpcTimeout      time.Duration
	BalanceMod      int8
	RegisterTimeout time.Duration
	UserName        string   //注册中心用户名
	Password        string   //注册中心密码
	Endpoints       []string //注册中心地址
	TTl             int64
}

func NewRpcConfig(globalConf *configs.Config) (*RpcConfig, error) {

	rpcConf := &RpcConfig{
		Cluster:         globalConf.GrpcService.Cluster,
		RpcPort:         globalConf.GrpcService.Port,
		Endpoints:       globalConf.Registry.EndPoints,
		RegisterTimeout: time.Duration(globalConf.Registry.RegisterTimeout) * time.Second,
		BalanceMod:      globalConf.GrpcService.BalanceMod,
		RpcTimeout:      time.Duration(globalConf.GrpcService.RpcTimeout) * time.Second,
		UserName:        globalConf.Registry.UserName,
		Password:        globalConf.Registry.Password,
		TTl:             globalConf.Registry.TTl,
	}

	return rpcConf, rpcConf.check()
}

func (c *RpcConfig) check() error {
	if c.RpcPort == "" {
		return fmt.Errorf("端口号不能为空")
	}

	if len(c.Endpoints) == 0 {
		return fmt.Errorf("服务中心集群地址不能为空")
	}

	return nil
}
