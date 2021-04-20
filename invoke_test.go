package main

import (
	"eapp/eapp"
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	app := eapp.NewApp()
	res, err := app.GetContainer().GetRpcInvoker().
		Request("",
			"yal-test",
			map[string]string{},
			map[string]interface{}{
				"id":   1,
				"name": "yyy",
			}, true)
	if err != nil {
		app.GetContainer().Errorf("err::%v", err)
	}
	res2, err := app.GetContainer().GetRpcInvoker().
		Request("default",
			"yal-test2",
			map[string]string{},
			map[string]interface{}{
				"id":   2,
				"name": "yyy2",
			}, true)
	fmt.Println("res:::::", res)
	fmt.Println("res2:::::", res2)
}
