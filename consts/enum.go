package consts

import "google.golang.org/grpc/naming"

const Modify naming.Operation = 0xFF //扩展naming.Operation

type ServerType uint8

const (
	HttpServer ServerType = iota + 1
	RpcServer
	MqcServer
	CronServer
)

func (s ServerType) String() string {
	switch s {
	case HttpServer:
		return "http"
	case RpcServer:
		return "grpc"
	case MqcServer:
		return "mqc"
	case CronServer:
		return "cron"
	}

	return ""
}
