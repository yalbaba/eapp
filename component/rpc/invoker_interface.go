package rpc

type RpcInvoker interface { //服务端要实现的RPC调用程序组件
	Request(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error)
	Close()
}
