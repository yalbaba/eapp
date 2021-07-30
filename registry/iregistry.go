package registry

import (
	"google.golang.org/grpc/naming"
)

type IRegistry interface { //客户端要实现的注册中心接口
	Register(service string, update naming.Update) (err error)
	Close() error
	Resolve(target string) (naming.Watcher, error)
}
