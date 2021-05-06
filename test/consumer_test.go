package test_test

import (
	"eapp/app"
	"testing"
)

func TestMqc(t *testing.T) {
	myapp := app.NewApp(
		app.WithMqcServer(),
	)

	myapp.GetContainer().Send("yangal", map[string]interface{}{
		"name": "yangal",
	})
}
