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
	EndPoints       []string `toml:"end_points"`
	UserName        string   `toml:"user_name"`
	Password        string   `toml:"password"`
	RegisterTimeout int      `toml:"register_timeout"`
	TTl             int64    `toml:"ttl"`
}

//grpc配置
type GrpcService struct {
	Cluster    string `toml:"cluster"`
	Port       string `toml:"port"`
	RpcTimeout int    `toml:"rpc_timeout"`
	BalanceMod int8   `toml:"balance_mod"`
}
