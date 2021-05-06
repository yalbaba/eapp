package rpc

import (
	"context"
	"eapp/pb"
	"encoding/json"
	"fmt"
)

//---------------接口------------------
type IRequest interface {
	Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error)
}

//---------------接口------------------

type RequestService struct {
	Servers map[string]IRequest
}

func (r *RequestService) Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error) {
	if _, ok := r.Servers[in.Service]; !ok {
		return &pb.ResponseContext{
			Status: 500,
			Result: "服务未注册",
		}, fmt.Errorf("服务未注册")
	}

	header := make(map[string]string)
	if err := json.Unmarshal([]byte(in.Header), &header); err != nil {
		return nil, err
	}

	input := make(map[string]interface{})
	if err := json.Unmarshal([]byte(in.Input), &input); err != nil {
		return nil, err
	}
	resp, err := r.Servers[in.Service].Handle(ctx, header, input)
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
