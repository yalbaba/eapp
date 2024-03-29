package global_config

var (
	Conf = &Config{}
)

type Config struct {
	Registry    *Registry    `toml:"registry"`
	GrpcService *GrpcService `toml:"grpc_service"`
	MqcService  *MqcService  `toml:"mqc_service"`
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

//mqc配置
type MqcService struct {
	Host string `toml:"host"`
}
