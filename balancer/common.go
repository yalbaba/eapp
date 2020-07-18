package balancer

import "google.golang.org/grpc"

type AddrInfo struct {
	Addr            grpc.Address
	Connected       bool
	Weight          int
	CurrentWeight   int
	EffectiveWeight int
}

//根据协议选项来决定使用哪种负载均衡算法
