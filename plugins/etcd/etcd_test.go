package etcd

import (
	"context"
	etcd "github.com/coreos/etcd/clientv3"
	"testing"
)

func TestRegistry(t *testing.T) {
	cli, err := etcd.NewFromURL("http://127.0.0.1:2379")
	if err != nil {
		t.Error(err)
	}
	cli.KV.Put(context.Background(), "ytest", "yal")
	//r, _ := cli.KV.Get(context.Background(), "ytest")
}
