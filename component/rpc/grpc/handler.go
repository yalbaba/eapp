package grpc

import (
	"context"
)

type RpcHandler interface { //客户端实现的rpc处理请求接口
	Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error)
}
