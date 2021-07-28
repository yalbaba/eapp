package servers

import (
	"eapp/component"
	"eapp/consts"
)

//多种服务的接口
type IServer interface {
	Start() error
	Stop()
	RegisterService(domain string, h interface{}) error //domain:任意的服务对象名称，rpc_server:服务名  mqc_server:topic名
}

type IServerResolver interface {
	Resolve(c component.Container) IServer
}

var resolvers = make(map[consts.ServerType]IServerResolver)

func Register(serverType consts.ServerType, resolver IServerResolver) {

	if _, ok := resolvers[serverType]; ok {
		panic("resolver exist server:" + serverType.String())
	}

	resolvers[serverType] = resolver
}

func NewServer(serverType consts.ServerType, c component.Container) IServer {

	if _, ok := resolvers[serverType]; ok {
		return resolvers[serverType].Resolve(c)
	}

	panic("resolver not exist server:" + serverType.String())
}
