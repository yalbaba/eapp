package test_test

import (
	"eapp/app"
	"fmt"
	"testing"
)

func TestRpc(t *testing.T) {
	app := app.NewApp()
	res, err := app.GetContainer().RpcRequest("",
		"yal-test",
		map[string]string{},
		map[string]interface{}{
			"id":   1,
			"name": "yyy",
		}, true)
	if err != nil {
		app.GetContainer().Errorf("err::%v", err)
	}
	res2, err := app.GetContainer().RpcRequest("default",
		"yal-test2",
		map[string]string{},
		map[string]interface{}{
			"id":   2,
			"name": "yyy2",
		}, true)
	fmt.Println("res:::::", res)
	fmt.Println("res2:::::", res2)
}
