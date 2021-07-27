package mqc

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"sync"
)

type MqcInvoker interface { //服务端要实现的mqc组件
	Publish(topic string, body []byte) error
	Close()
}

//---------以下是实现-----------
type NsqInvoker struct {
	sync.RWMutex
	producer *nsq.Producer
}

func NewNsqInvoker(host string) *NsqInvoker {
	p, err := nsq.NewProducer(host, nsq.NewConfig())
	if err != nil {
		panic(fmt.Errorf("创建nsq服务失败,err:%v", err))
	}
	return &NsqInvoker{
		producer: p,
	}
}

func (m *NsqInvoker) Publish(topic string, body []byte) error {
	return m.producer.Publish(topic, body)
}

func (m *NsqInvoker) Close() {
	m.RLock()
	defer m.RUnlock()
	m.producer.Stop()
}
