package example

import (
	"erpc/configs"
	"erpc/servers"
	"fmt"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	rpcConfs, err := servers.NewRpcConfig(configs.Conf,
		servers.WithTimeOut(time.Second),
		servers.WithBalancer(11))
	if err != nil {
		fmt.Println("NewRpcConfig :::", err)
		return
	}

	eserver, err := servers.NewErpcServer(rpcConfs)
	if err != nil {
		fmt.Println("NewErpcServer:::", err)
		return
	}

	resp, err := eserver.Rpc("yal-test", map[string]interface{}{
		"id":   1,
		"name": "yang",
	})

	fmt.Println("resp:::::::::::", resp)

	resp2, err := eserver.Rpc("yal-test2", map[string]interface{}{
		"id":   2,
		"name": "yang2",
	})

	fmt.Println("resp2::::::::::", resp2)

}
