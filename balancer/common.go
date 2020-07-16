package balancer

import "google.golang.org/grpc"

type AddrInfo struct {
	Addr            grpc.Address
	Connected       bool
	Weight          int
	CurrentWeight   int
	EffectiveWeight int
}
