package server

import (
	"context"
)

type Handler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

type IServer interface {
	Start() error
	Stop()
	RegistService(serviceName string, h Handler) error
	Rpc(cluster, service string, header map[string]interface{}, input map[string]interface{}) (interface{}, error)
}
