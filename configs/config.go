package configs

import (
	"github.com/BurntSushi/toml"
)

var (
	Conf = &Config{}
)

func init() {
	_, err := toml.DecodeFile("./configs/config.toml", Conf)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	Registry    *Registry    `toml:"registry"`
	GrpcService *GrpcService `toml:"grpc_service"`
}

//注册中心配置对象
type Registry struct {
	Addrs           []string `toml:"addrs"`
	UserName        string   `toml:"user_name"`
	Password        string   `toml:"password"`
	RegisterTimeOut int      `toml:"register_time_out"`
	TTl             int      `toml:"t_tl"`
}

//grpc配置
type GrpcService struct {
	Cluster    string `toml:"cluster"`
	Port       string `toml:"port"`
	RpcTimeOut int    `toml:"rpc_time_out"`
}
