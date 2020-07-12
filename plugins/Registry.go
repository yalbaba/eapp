package plugins

import (
	"context"
	"google.golang.org/grpc/naming"
)

type Registry interface {
	Register(ctx context.Context, cluster, service string, update naming.Update) (err error)
	Close()
	Resolve(target string) (naming.Watcher, error)
}
