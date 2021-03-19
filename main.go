package main

import (
	"context"
	"erpc/server"
	"fmt"
	"time"
)

func main() {
	eserver, err := server.NewApp(server.WithTimeOut(time.Second),
		server.WithBalancer(11))
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	resp, err := eserver.Rpc("default", "yal-test", map[string]string{}, map[string]interface{}{
		"id":   1,
		"name": "yang",
	}, true)

	fmt.Println("resp:::::::::::", resp)

	resp2, err := eserver.Rpc("default", "yal-test2", map[string]string{}, map[string]interface{}{
		"id":   2,
		"name": "yang2",
	}, true)

	fmt.Println("resp2::::::::::", resp2)
}

func MyHandler(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务1")
	//进行链路追踪使用header
	return fmt.Sprintf("input1:::%+v", input), nil
}

func MyHandler2(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务2")
	//进行链路追踪使用header
	return fmt.Sprintf("input2:::%+v", input), nil
}
