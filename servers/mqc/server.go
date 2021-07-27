package mqc

import (
	"eapp/component"
	"eapp/component/mqc"
	"eapp/consts"
	"eapp/global_config"
	"eapp/servers"
	"github.com/nsqio/go-nsq"
	"strings"
	"sync"
)

type MqcServer struct {
	services  map[string]mqc.MqcHandler
	consumers []*nsq.Consumer
	c         component.Container
	lock      sync.Mutex
}

func NewMqcServer(c component.Container) *MqcServer {
	return &MqcServer{
		c:        c,
		services: make(map[string]mqc.MqcHandler),
	}
}

func (m *MqcServer) Start() error {

	m.c.Debug("MQC服务器正在启动...")

	for k, v := range m.services {
		arr := strings.Split(k, "/")
		m.lock.Lock()
		c, err := nsq.NewConsumer(arr[0], arr[1], nsq.NewConfig()) // 新建一个消费者
		if err != nil {
			panic(err)
		}

		//添加nsq的消费者模块
		c.AddHandler(v)

		if err := c.ConnectToNSQD(global_config.Conf.MqcService.Host); err != nil { // 建立连接
			panic(err)
		}
		m.consumers = append(m.consumers, c)
		m.lock.Unlock()
	}

	m.c.Debug("MQC服务器启动成功...")
	return nil
}

func (m *MqcServer) Stop() {
	for _, v := range m.consumers {
		v.Stop()
	}
	return
}

//service = "topic_name/channel_name"
func (m *MqcServer) RegisterService(service string, h interface{}) error {

	arr := strings.Split(service, "/")
	if service == "" || len(arr) < 2 || arr[0] == "" || arr[1] == "" {
		panic("服务名错误")
		return nil
	}

	if _, ok := h.(mqc.MqcHandler); !ok {
		panic("消息服务类型错误")
		return nil
	}

	m.services[service] = h.(mqc.MqcHandler)

	return nil
}

//适配器
type mqcServerAdapter struct{}

func (*mqcServerAdapter) Resolve(c component.Container) servers.IServer {

	mqcServer := NewMqcServer(c)

	return mqcServer
}

func init() {
	servers.Register(consts.MqcServer, &mqcServerAdapter{})
}
