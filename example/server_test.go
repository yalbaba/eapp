package example

import (
	"context"
	"erpc/global"
	"erpc/servers"
	"fmt"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	rpcConfs, err := servers.NewRpcConfig(&global.GlobalConfig{
		Cluster:       "yal",
		Port:          "9091",
		RegisterAddrs: []string{"127.0.0.1:2379"},
	}, servers.WithTimeOut(time.Second), servers.WithBalancer(11))
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

	resp, err := eserver.Rpc("yal-test", map[string]interface{}{
		"id":   1,
		"name": "123",
	})
	if err != nil {
		fmt.Println("err::", err)
		return
	}

	fmt.Println("resp::::", resp)

	resp2, err := eserver.Rpc("yal-test2", map[string]interface{}{
		"id":   1,
		"name": "123",
	})
	if err != nil {
		fmt.Println("err::", err)
		return
	}

	fmt.Println("resp2::::", resp2)

	time.Sleep(time.Second)
}

func MyHandler(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务1")
	return fmt.Sprintf("input1:::%+v", input), nil
}

func MyHandler2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务2")
	return fmt.Sprintf("input2:::%+v", input), nil
}