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

	resp, err := eserver.Rpc("default", "yal-test", map[string]interface{}{
		"id":   1,
		"name": "yang",
	})

	fmt.Println("resp:::::::::::", resp)

	resp2, err := eserver.Rpc("default", "yal-test2", map[string]interface{}{
		"id":   2,
		"name": "yang2",
	})

	fmt.Println("resp2::::::::::", resp2)
}

func MyHandler(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务1")
	return fmt.Sprintf("input1:::%+v", input), nil
}

func MyHandler2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务2")
	return fmt.Sprintf("input2:::%+v", input), nil
}
