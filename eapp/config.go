package eapp

import (
	"eapp/consts"
	"eapp/logger"
)

type AppConfigs struct {
	ServerTypes map[consts.ServerType]bool
	Log         logger.ILogger
}

type Option func(conf *AppConfigs)

func WithRpcServer() Option {
	return func(conf *AppConfigs) {
		conf.ServerTypes[consts.RpcServer] = true
	}
}

func WithLogger(l logger.ILogger) Option {
	return func(o *AppConfigs) {
		o.Log = l
	}
}
