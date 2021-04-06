package server

import (
	"context"
	"eapp/pb"
	"encoding/json"
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

	header := make(map[string]string)
	if err := json.Unmarshal([]byte(in.Header), &header); err != nil {
		return nil, err
	}

	input := make(map[string]interface{})
	if err := json.Unmarshal([]byte(in.Input), &input); err != nil {
		return nil, err
	}
	resp, err := r.servers[in.Service](ctx, header, input)
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
