package server

import (
	"erpc/configs"
	"erpc/logger"
	"erpc/logger/zap"
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
	RpcTimeOut      time.Duration
	BalancerMod     int8
	RegisterTimeOut time.Duration
	UserName        string   //注册中心用户名
	Password        string   //注册中心密码
	RegisterAddrs   []string //注册中心地址
	TTl             int64
	log             logger.ILogger
}

func NewRpcConfig(conf *configs.Config, opts ...option) (*RpcConfig, error) {
	opt := &RpcConfigOptions{}
	for _, o := range opts {
		o(opt)
	}

	if opt.log == nil {
		opt.log = zap.NewDefaultLogger(&logger.LoggerConfig{
			OutputDir: "./logs/",
			OutFile:   "default.log",
			ErrFile:   "default.err",
		})
	}

	rpcConf := &RpcConfig{
		Cluster:         conf.GrpcService.Cluster,
		RpcPort:         conf.GrpcService.Port,
		RegisterAddrs:   conf.Registry.Addrs,
		RegisterTimeOut: opt.RegisterTimeOut,
		BalancerMod:     opt.BalancerMod,
		RpcTimeOut:      opt.RpcTimeOut,
		UserName:        conf.Registry.UserName,
		Password:        conf.Registry.Password,
		TTl:             opt.TTl,
		log:             opt.log,
	}

	return rpcConf, rpcConf.check()
}

func (c *RpcConfig) check() error {
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
	TTl             int64
	log             logger.ILogger
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

func WithTTl(ttl int64) option {
	return func(o *RpcConfigOptions) {
		o.TTl = ttl
	}
}

func WithLogger(l logger.ILogger) option {
	return func(o *RpcConfigOptions) {
		o.log = l
	}
}
