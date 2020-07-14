package main

import (
	"erpc/plugins/logger"
	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLoggerV2(logger.NewZapLogger())
}
