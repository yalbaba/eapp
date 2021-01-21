package servers

import (
	"context"
	"encoding/json"
	"erpc/iservers"
	"erpc/pb"
)

type RequestService struct {
	handle iservers.Handler
}

func (r *RequestService) Request(ctx context.Context, in *pb.RequestContext) (*pb.ResponseContext, error) {
	input := make(map[string]interface{})
	json.Unmarshal([]byte(in.Input), &input)
	resp, err := r.handle(ctx, input)
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
