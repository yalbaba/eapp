package servers

import (
	"context"
	"erpc/iservers"
	"erpc/pb"
)

type RequestService struct {
	input  map[string]interface{} //参数
	handle iservers.Handler
}

func (r *RequestService) GetService() string {
	//todo
	return ""
}

func (r *RequestService) GetMethod() string {
	//todo
	return ""
}

func (r *RequestService) GetForm() map[string]interface{} {
	return r.input
}

func (r *RequestService) Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error) {
	resp, err := r.handle(ctx, r.input)
	if err != nil {
		return &pb.ResponseContext{
			Status: 500,
			Result: resp.(string),
		}, err
	}

	return &pb.ResponseContext{
		Status: 500,
		Result: resp.(string),
	}, nil
}
