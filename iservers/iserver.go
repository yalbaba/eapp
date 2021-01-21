package iservers

import (
	"context"
)

type Handler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

type IServer interface {
	Start() error
	Run() error
	Stop()
	RegistService(serviceName string, h Handler) error
	Rpc(serviceName string, input map[string]interface{}) (interface{}, error)
}
