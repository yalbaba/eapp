package server

import (
	"context"
)

type Handler func(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error)

type IServer interface {
	Start() error
	Stop()
	RegisterService(service string, h Handler) error
	Rpc(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error)
}
