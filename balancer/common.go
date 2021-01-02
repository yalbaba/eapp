package balancer

import "google.golang.org/grpc"

//每个服务的集群子节点地址信息
type AddrInfo struct {
	Addr            grpc.Address //服务地址
	Connected       bool         //连接状态
	Weight          int          //总权重
	CurrentWeight   int          //当前权重
	EffectiveWeight int          //有效权重
}

//根据协议选项来决定使用哪种负载均衡算法
