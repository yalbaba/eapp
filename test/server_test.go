package test

import (
	"context"
	"eapp/app"
	"eapp/component"
	"eapp/component/rpc"
	"fmt"
	"github.com/nsqio/go-nsq"
	"testing"
)

func TestServer(t *testing.T) {
	app := app.NewApp(
		app.WithRpcServer(),
		app.WithMqcServer(),
	)

	//err := app.RegisterRpcService("yal-test", NewHandler(app.GetContainer()))
	//if err != nil {
	//	panic(err)
	//}
	//app.RegisterRpcService("yal-test2", NewHandler2(app.GetContainer()))
	app.RegisterMqcService("yangal", "cha1", NewMqcHandler(app.GetContainer()))
	app.Start()
}

type testMqcHandler struct {
	c component.Container
}

func NewMqcHandler(c component.Container) *testMqcHandler {
	return &testMqcHandler{
		c: c,
	}
}

func (t *testMqcHandler) HandleMessage(message *nsq.Message) error {
	t.c.Warn("msg:::", string(message.Body))
	return nil
}

type testhandler struct {
	c component.Container
}

type testhandler2 struct {
	c component.Container
}

func NewHandler(c component.Container) rpc.IRequest {
	return &testhandler{c: c}
}

func NewHandler2(c component.Container) rpc.IRequest {
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
