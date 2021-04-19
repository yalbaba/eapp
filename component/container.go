package component

import (
	"eapp/component/rpc"
	"eapp/logger"
)

//给各个服务器使用的公共组件
type Container interface {
	logger.ILogger
}

//给app使用的公共组件
type IComponent interface {
	GetRpcInvoker() rpc.RpcInvoker
	logger.ILogger
	Close()
}

type component struct {
	Rpc rpc.RpcInvoker
	logger.ILogger
}

func NewComponent(r rpc.RpcInvoker, l logger.ILogger) IComponent {
	c := component{Rpc: r}
	c.ILogger = l
	return c
}

func (c component) GetRpcInvoker() rpc.RpcInvoker {
	return c.Rpc
}

func (c component) Close() {
	//关闭所有公共组件
}
