package iservers

type IServer interface {
	Start() error
	Run() error
	Stop()
	RegistService(cluster, serviceName string, addrs []string, h IHandler) error
}
