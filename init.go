package main

import (
	"erpc/config"
	"google.golang.org/grpc/grpclog"
)

func init() {
	logConf := config.LoggerConfig{}
	grpclog.SetLoggerV2(logConf.NewLogger())
}
