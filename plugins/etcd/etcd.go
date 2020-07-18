package etcd

import (
	"context"
	"encoding/json"
	"erpc/plugins"
	"fmt"
	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"
	"time"
)

const (
	resolverTimeOut = 10 * time.Second
)

//服务发现相关
type etcdWatcher struct {
	cli       *etcd.Client
	target    string //格式：cluster/service
	cancel    context.CancelFunc
	ctx       context.Context
	watchChan etcd.WatchChan
}

func (w *etcdWatcher) Next() ([]*naming.Update, error) {
	var updates []*naming.Update
	//注册中心第一次开启监控
	if w.watchChan != nil {
		resp, err := w.cli.Get(context.Background(), w.target)
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
		//etcd.WithSerializable() 暂时未搞懂
		opt := []etcd.OpOption{etcd.WithRev(resp.Header.Revision + 1), etcd.WithPrefix(), etcd.WithPrevKV()}
		w.watchChan = w.cli.Watch(context.TODO(), w.target, opt...)
		return updates, nil
	}
	//当发生改变，会得到watch通知
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
		case etcd.EventTypePut:
			err = json.Unmarshal(event.Kv.Value, &update)
			update.Op = naming.Add
		case etcd.EventTypeDelete:
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

func (w *etcdWatcher) Close() {
	w.cli.Close()
	w.cancel()
}

//注册中心相关
type etcdRegistry struct {
	cli    *etcd.Client
	lci    etcd.Lease
	cancal context.CancelFunc
}

/*
target = cluster+/+service
*/
func (r *etcdRegistry) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), resolverTimeOut)
	watcher := &etcdWatcher{
		cli:    r.cli,
		target: target,
		ctx:    ctx,
		cancel: cancel,
	}
	return watcher, nil
}

func NewEtcdRegistry(c *etcd.Client) plugins.Registry {
	return &etcdRegistry{
		cli: c,
		lci: etcd.NewLease(c),
	}
}

//注册服务
//服务格式 key: cluster/service/address value: 包括服务的地址在内的其他元数据
func (r *etcdRegistry) Register(ctx context.Context, cluster, service string, update naming.Update) (err error) {
	var upBytes []byte
	upBytes, err = json.Marshal(&update)
	if err != nil {
		status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, cancal := context.WithTimeout(context.TODO(), resolverTimeOut)
	r.cancal = cancal

	key := cluster + "/" + service + "/" + update.Addr
	switch update.Op {
	case naming.Add:
		lRsp, err := r.lci.Grant(ctx, 100) //这个服务租约时间需要封装在配置对象中
		if err != nil {
			return err
		}
		opts := []etcd.OpOption{etcd.WithLease(lRsp.ID)}
		r.cli.KV.Put(ctx, key, string(upBytes), opts...)
		if err != nil {
			return fmt.Errorf("注册服务异常,err:%v", err)
		}
		grpclog.Infof("etcd put key:%v value:%v\n", key, string(upBytes))
		lsRspChan, err := r.lci.KeepAlive(context.TODO(), lRsp.ID)
		if err != nil {
			return err
		}
		//	续租约直到服务挂掉
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

//关闭注册中心
func (r *etcdRegistry) Close() {
	r.cancal()
	r.lci.Close()
	r.cli.Close()
}
