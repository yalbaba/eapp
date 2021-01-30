package main

import (
	"context"
	"erpc/configs"
	"erpc/servers"
	"fmt"
	"time"
)

func main() {
	rpcConfs, err := servers.NewRpcConfig(configs.Conf, servers.WithTimeOut(time.Second), servers.WithBalancer(11))
	if err != nil {
		fmt.Println("NewRpcConfig :::", err)
		return
	}

	eserver, err := servers.NewErpcServer(rpcConfs)
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	err = eserver.RegistService("yal-test", MyHandler)
	if err != nil {
		fmt.Println("RegistService:::", err)
		return
	}
	err = eserver.RegistService("yal-test2", MyHandler2)
	if err != nil {
		fmt.Println("RegistService:::", err)
		return
	}

	err = eserver.Start()
	if err != nil {
		fmt.Println("Start:::", err)
		return
	}
}

func MyHandler(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务1")
	return fmt.Sprintf("input1:::%+v", input), nil
}

func MyHandler2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务2")
	return fmt.Sprintf("input2:::%+v", input), nil
}
