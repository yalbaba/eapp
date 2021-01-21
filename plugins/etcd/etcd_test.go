package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	cli, err := clientv3.New(conf)
	if err != nil {
		fmt.Println("err::", err)
		return
	}
	res2, err := cli.KV.Put(context.Background(), "ytest", "yal")
	if err != nil {
		fmt.Println("err::", err)
		return
	}
	fmt.Println("res2:::", res2.PrevKv)
	res, err := cli.KV.Get(context.Background(), "ytest")
	if err != nil {
		fmt.Println("err::", err)
		return
	}
	fmt.Println("res:::", res.Kvs)
	cli.Close()
}
