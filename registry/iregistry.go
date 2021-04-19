package registry

import (
	"google.golang.org/grpc/naming"
)

type IRegistry interface {
	Register(service string, update naming.Update) (err error)
	Close() error
	Resolve(target string) (naming.Watcher, error)
}
