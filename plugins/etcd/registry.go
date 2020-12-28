package etcd

import (
	"erpc/plugins"
	etcdv3 "github.com/coreos/etcd/clientv3"
)

//后面用适配器模式实现,注册中心配置对象
type RegistryConfig struct {
	RegistryType string   `yaml:"registry_type"` //etcd default
	Addr         []string `yaml:"endpoints"`
	UserName     string   `yaml:"user_name"`
	Pass         string   `yaml:"pass"`
}

func (conf *RegistryConfig) NewRegistry() (registry plugins.Registry, err error) {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: conf.Addr,
		Username:  conf.UserName,
		Password:  conf.Pass,
	})
	return NewEtcdRegistry(cli), nil
}
