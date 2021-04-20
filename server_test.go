package main

import (
	"context"
	"eapp/component"
	"eapp/component/rpc"
	"eapp/eapp"
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	app := eapp.NewApp(
		eapp.WithRpcServer(),
	)

	err := app.Rpc("yal-test", NewHandler(app.GetContainer()))
	if err != nil {
		panic(err)
	}
	app.Rpc("yal-test2", NewHandler2(app.GetContainer()))
	app.Start()
}

type testhandler struct {
	c component.Container
}

type testhandler2 struct {
	c component.Container
}

func NewHandler(c component.Container) rpc.IHandler {
	return &testhandler{c: c}
}

func NewHandler2(c component.Container) rpc.IHandler {
	return &testhandler2{c: c}
}

func (t *testhandler) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	t.c.Warn("执行任务1")
	//进行链路追踪使用header
	return fmt.Sprintf("input1:::%+v", input), nil
}

func (t *testhandler2) Handle(ctx context.Context, header map[string]string, input map[string]interface{}) (interface{}, error) {
	t.c.Warn("执行任务2")
	//进行链路追踪使用header
	return fmt.Sprintf("input2:::%+v", input), nil
}
