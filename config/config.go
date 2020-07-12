package config

import (
	"erpc/plugins"
	"erpc/plugins/retcd"
	etcd "github.com/coreos/etcd/clientv3"
)

//后面用适配器模式实现
type RegistryConfig struct {
	RegistryType string   `yaml:"registry_type"` //retcd default
	Addr         []string `yaml:"endpoints"`
	UserName     string   `yaml:"user_name"`
	Pass         string   `yaml:"pass"`
}

func (conf *RegistryConfig) NewRegistry() (registry plugins.Registry, err error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints: conf.Addr,
		Username:  conf.UserName,
		Password:  conf.Pass,
	})
	return retcd.NewEtcdRegistry(cli), nil
}
