package servers

import (
	"context"
	"encoding/json"
	"erpc/pb"
)

type RequestService struct {
	servers map[string]Handler
}

func (r *RequestService) Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error) {
	if _, ok := r.servers[in.Service]; !ok {
		return &pb.ResponseContext{
			Status: 500,
			Result: "服务未注册",
		}, nil
	}

	input := make(map[string]interface{})
	json.Unmarshal([]byte(in.Input), &input)
	resp, err := r.servers[in.Service](ctx, input)
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
