package balancer

import (
	"context"
	"google.golang.org/grpc"
)

type IBalancer interface {
	Start(target string, config grpc.BalancerConfig) error
	Up(addr grpc.Address) func(error)
	Get(ctx context.Context, opts grpc.BalancerGetOptions) (addr grpc.Address, put func(), err error)
	Notify() <-chan []grpc.Address
	Close() error
}

//每个服务的集群子节点地址信息
type AddrInfo struct {
	Addr            grpc.Address //服务地址
	Connected       bool         //连接状态
	Weight          int          //总权重
	CurrentWeight   int          //当前权重
	EffectiveWeight int          //有效权重
}
