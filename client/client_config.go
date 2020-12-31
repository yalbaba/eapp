package client

import "time"

const (
	RoundRobin = iota + 10
	Random
)

type ClientConfig struct {
	RpcTimeOut      time.Time
	BalancerMod     int8
	RegisterTimeOut time.Time
}

type ClientOptions struct {
	RpcTimeOut      time.Time
	BalancerMod     int8
	RegisterTimeOut time.Time
}

type option func(o *ClientOptions)

func WithTimeOut(t time.Time) option {
	return func(o *ClientOptions) {
		o.RpcTimeOut = t
	}
}

func WithBalancer(mod int8) option {
	return func(o *ClientOptions) {
		o.BalancerMod = mod
	}
}

func WithRegisterTimeOut(t time.Time) option {
	return func(o *ClientOptions) {
		o.RegisterTimeOut = t
	}
}
