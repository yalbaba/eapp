package nsq

import (
	"eapp/component"
	app_nsq "eapp/component/mqc/nsq"
	"eapp/consts"
	"eapp/global_config"
	"eapp/servers"
	"github.com/nsqio/go-nsq"
	"strings"
	"sync"
)

//接入nsq
type MqcServer struct {
	services  map[string]app_nsq.MqcHandler
	consumers []*nsq.Consumer
	c         component.Container
	lock      sync.Mutex
}

func NewMqcServer(c component.Container) *MqcServer {
	return &MqcServer{
		c:        c,
		services: make(map[string]app_nsq.MqcHandler),
	}
}

func (m *MqcServer) Start() error {

	m.c.Debug("MQC服务器正在启动...")

	for k, v := range m.services {
		arr := strings.Split(k, "/")
		m.lock.Lock()
		c, err := nsq.NewConsumer(arr[0], arr[1], nsq.NewConfig())
		if err != nil {
			panic(err)
		}

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
func (m *MqcServer) RegisterService(topic string, h interface{}) error {

	arr := strings.Split(topic, "/")
	if topic == "" || len(arr) < 2 || arr[0] == "" || arr[1] == "" {
		panic("服务名错误")
		return nil
	}

	if _, ok := h.(app_nsq.MqcHandler); !ok {
		panic("消息服务类型错误")
		return nil
	}

	m.services[topic] = h.(app_nsq.MqcHandler)

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
