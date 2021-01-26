package servers

import (
	"erpc/configs"
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
	RegisterAddrs   []string //注册中心地址
	RpcTimeOut      time.Duration
	BalancerMod     int8
	RegisterTimeOut time.Duration
	UserName        string //服务中心用户名
	Pass            string //服务中心密码
}

func NewRpcConfig(conf *configs.Config, opts ...option) (*RpcConfig, error) {
	opt := &RpcConfigOptions{}
	for _, o := range opts {
		o(opt)
	}

	rpcConf := &RpcConfig{
		Cluster:         conf.GrpcService.Cluster,
		RpcPort:         conf.GrpcService.Port,
		RegisterAddrs:   conf.Registry.Addrs,
		RegisterTimeOut: opt.RegisterTimeOut,
		BalancerMod:     opt.BalancerMod,
		RpcTimeOut:      opt.RpcTimeOut,
		UserName:        conf.Registry.UserName,
		Pass:            conf.Registry.Password,
	}

	return rpcConf, rpcConf.check()
}

func (c *RpcConfig) check() error {
	if c.RegisterTimeOut == 0 {
		c.RegisterTimeOut = 5 * time.Second
	}

	if c.RpcTimeOut == 0 {
		c.RpcTimeOut = 3 * time.Second
	}

	if c.RpcPort == "" {
		return fmt.Errorf("端口号不能为空")
	}

	if len(c.RegisterAddrs) == 0 {
		return fmt.Errorf("服务中心集群地址不能为空")
	}

	return nil
}

type RpcConfigOptions struct {
	RpcTimeOut      time.Duration
	BalancerMod     int8
	RegisterTimeOut time.Duration
}

type option func(o *RpcConfigOptions)

func WithTimeOut(t time.Duration) option {
	return func(o *RpcConfigOptions) {
		o.RpcTimeOut = t
	}
}

func WithBalancer(mod int8) option {
	return func(o *RpcConfigOptions) {
		o.BalancerMod = mod
	}
}

func WithRegisterTimeOut(t time.Duration) option {
	return func(o *RpcConfigOptions) {
		o.RegisterTimeOut = t
	}
}
