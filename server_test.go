package main

import (
	"context"
	"eapp/eapp"
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	app := eapp.NewApp(
		eapp.WithRpcServer(),
	)

	err := app.Rpc("yal-test", &testhandler{})
	if err != nil {
		panic(err)
	}
	app.Rpc("yal-test2", &testhandler2{})
	app.Start()
}

type testhandler struct {
}

type testhandler2 struct {
}

func (*testhandler) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务1")
	//进行链路追踪使用header
	return fmt.Sprintf("input1:::%+v", input), nil
}

func (*testhandler2) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	fmt.Println("执行rpc任务2")
	//进行链路追踪使用header
	return fmt.Sprintf("input2:::%+v", input), nil
}
