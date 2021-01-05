package servers

import "erpc/global"

//rpc服务的配置
type RpcConfig struct {
	Cluster  string
	Services map[string]string //rpc服务和地址
	RpcAddr  string
	RpcPort  string
}

func NewRpcConfig(conf *global.ServerConfig) *RpcConfig {
	return &RpcConfig{
		Services: make(map[string]string),
		Cluster:  "default",
		RpcPort:  conf.RpcPort,
		RpcAddr:  conf.RpcAddr,
	}
}
