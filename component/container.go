package component

import "eapp/logger"

//给各个服务器使用的公共组件
type Container interface {
	logger.ILogger
}

//给app使用的公共组件
type IComponent interface {
	logger.ILogger
	Close()
}

type component struct {
	logger.ILogger
}

func (c component) Close() {
	//关闭所有公共组件
}

func NewComponent(l logger.ILogger) IComponent {
	return component{l}
}
