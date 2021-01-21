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
		RegisterAddrs: []string{"127.0.0.1:2379"},
	}, "9091", servers.WithTimeOut(time.Second))
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

}

func MyHandler(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务")
	return fmt.Sprintf("input:::%+v", input), nil
}
