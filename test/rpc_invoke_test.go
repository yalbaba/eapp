package test_test

import (
	"eapp/app"
	"fmt"
	"testing"
)

func TestRpc(t *testing.T) {
	a := app.NewApp(
		app.WithRpcServer(),
	)
	res, err := a.GetContainer().RpcRequest("default",
		"yal-test",
		map[string]string{},
		map[string]interface{}{
			"id":   1,
			"name": "yyy",
		}, true)
	if err != nil {
		a.GetContainer().Errorf("err::%v", err)
	}
	res2, err := a.GetContainer().RpcRequest("default",
		"yal-test2",
		map[string]string{},
		map[string]interface{}{
			"id":   2,
			"name": "yyy2",
		}, true)
	fmt.Println("res:::::", res)
	fmt.Println("res2:::::", res2)
}
