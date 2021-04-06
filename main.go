package main

import (
	"context"
	"eapp/server"
	"fmt"
	"time"
)

func main() {
	app, err := server.NewApp(server.WithTimeOut(time.Second),
		server.WithBalancer(11))
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	app.RegisterService("yal-test", MyHandler)
	app.RegisterService("yal-test2", MyHandler2)
	app.Start()
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
