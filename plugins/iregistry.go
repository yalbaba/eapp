package plugins

import (
	"google.golang.org/grpc/naming"
)

type IRegistry interface {
	Register(cluster, service string, update naming.Update) (err error)
	Close()
	Resolve(target string) (naming.Watcher, error)
}
