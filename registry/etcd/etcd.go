package etcd

import (
	"context"
	"eapp/registry"
	"encoding/json"
	"fmt"
	etcdv3 "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"
	"time"
)

const (
	ResolverTimeOut = 10 * time.Second
)

//服务发现相关
type EtcdWatcher struct {
	cli       *etcdv3.Client
	target    string //格式：/cluster/service 服务的前缀名
	cancel    context.CancelFunc
	ctx       context.Context
	watchChan etcdv3.WatchChan
}

func (w *EtcdWatcher) Next() ([]*naming.Update, error) {

	//注册中心第一次开启监控
	if w.watchChan == nil {
		return w.firstNext()
	}

	var updates []*naming.Update
	// 开启监控后，当发生改变，会得到watch通知
	wrsp, ok := <-w.watchChan
	if !ok {
		err := status.Error(codes.Unavailable, "etcd watcher closed")
		return nil, err
	}
	if wrsp.Err() != nil {
		return nil, wrsp.Err()
	}
	//解析watch收到的事件
	for _, event := range wrsp.Events {
		var update naming.Update
		var err error
		switch event.Type {
		case etcdv3.EventTypePut:
			err = json.Unmarshal(event.Kv.Value, &update)
			update.Op = naming.Add
		case etcdv3.EventTypeDelete:
			err = json.Unmarshal(event.Kv.Value, &update)
			update.Op = naming.Delete
		}
		if err != nil {
			return nil, err
		}
		updates = append(updates, &update)
	}

	return updates, nil
}

//首次开启监控
func (w *EtcdWatcher) firstNext() ([]*naming.Update, error) {
	var updates []*naming.Update
	resp, err := w.cli.Get(context.Background(), w.target, etcdv3.WithPrefix()) //todo 忘了加
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	for _, kv := range resp.Kvs {
		var update naming.Update
		if err := json.Unmarshal(kv.Value, &update); err != nil {
			return nil, err
		}
		updates = append(updates, &update)
	}

	//etcd.WithPrefix(), key的前缀匹配
	opt := []etcdv3.OpOption{etcdv3.WithRev(resp.Header.Revision + 1), etcdv3.WithPrefix(), etcdv3.WithPrevKV()}
	w.watchChan = w.cli.Watch(context.TODO(), w.target, opt...)
	return updates, nil
}

func (w *EtcdWatcher) Close() {
	w.cli.Close()
	w.cancel()
}

//注册中心对象
type EtcdRegistry struct {
	cli    *etcdv3.Client
	lci    etcdv3.Lease
	cancel context.CancelFunc
	ttl    int64
}

func NewEtcdRegistry(cli *etcdv3.Client, ttl int64) registry.IRegistry {
	return &EtcdRegistry{
		cli: cli,
		lci: etcdv3.NewLease(cli),
		ttl: ttl,
	}
}

//注册服务
//服务格式 key: /cluster/service/192.168.0.2:9090 value: 包括服务的地址在内的其他元数据
func (r *EtcdRegistry) Register(service string, update naming.Update) (err error) {
	var upBytes []byte
	upBytes, err = json.Marshal(&update)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, c := context.WithTimeout(context.TODO(), ResolverTimeOut)
	r.cancel = c

	key := "/" + service + "/" + update.Addr
	switch update.Op {
	case naming.Add:
		lRsp, err := r.lci.Grant(ctx, r.ttl)
		if err != nil {
			return err
		}

		opts := []etcdv3.OpOption{etcdv3.WithLease(lRsp.ID)}
		_, err = r.cli.KV.Put(ctx, key, string(upBytes), opts...)
		if err != nil {
			return fmt.Errorf("注册服务异常,err:%v", err)
		}

		lsRspChan, err := r.lci.KeepAlive(context.Background(), lRsp.ID)
		if err != nil {
			return err
		}
		go func() {
			for {
				_, ok := <-lsRspChan
				if !ok {
					grpclog.Fatalf("%v 服务正在关闭", key)
					break
				}
			}
		}()
	case naming.Delete:
		_, err = r.cli.Delete(ctx, key)
		if err != nil {
			return err
		}
	default:
		return status.Error(codes.InvalidArgument, "不支持的操作")
	}
	return nil
}

func (r *EtcdRegistry) Close() error {
	r.cancel()
	err := r.lci.Close()
	if err != nil {
		return err
	}
	return r.cli.Close()
}

/*
target = /cluster/service
*/
func (r *EtcdRegistry) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ResolverTimeOut)
	watcher := &EtcdWatcher{
		cli:    r.cli,
		target: target,
		ctx:    ctx,
		cancel: cancel,
	}
	return watcher, nil
}
