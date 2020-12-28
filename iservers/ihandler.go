package iservers

import (
	"context"
	"erpc/pb"
)

type IHandler interface {
	Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error)
}
