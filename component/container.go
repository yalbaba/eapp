package component

import (
	"eapp/component/rpc"
	"eapp/global_config"
	"eapp/logger"
	"fmt"
	jsoniter "github.com/json-iterator/go"

	//jsoniter "github.com/json-iterator/go"
	"github.com/nsqio/go-nsq"
)

//给各个服务器使用的公共组件
type Container interface {
	RpcRequest(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error)
	Send(topic string, body map[string]interface{}) error

	logger.ILogger
}

//给app使用的公共组件
type IComponent interface {
	RpcRequest(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error)
	Send(topic string, body map[string]interface{}) error
	NewRpcInvoker()
	NewMqcProducer()

	logger.ILogger
	Close()
}

type component struct {
	RpcInvoker  rpc.RpcInvoker
	MqcProducer *nsq.Producer
	Num         int

	logger.ILogger
}

func NewComponent(l logger.ILogger) IComponent {
	c := &component{}
	c.ILogger = l
	return c
}

func (c *component) NewMqcProducer() {
	p, err := nsq.NewProducer(global_config.Conf.MqcService.Host, nsq.NewConfig())
	if err != nil {
		panic(fmt.Errorf("创建nsq服务失败,err:%v", err))
		return
	}
	c.MqcProducer = p
}

func (c *component) NewRpcInvoker() {
	c.RpcInvoker = rpc.NewRpcInvoker(global_config.Conf)
}

//Rpc调用
func (c *component) RpcRequest(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error) {
	return c.RpcInvoker.Request(cluster, service, header, input, failFast)
}

//发送消息
func (c *component) Send(topic string, body map[string]interface{}) error {
	b, err := jsoniter.Marshal(body)
	if err != nil {
		return fmt.Errorf("消息转json失败,err:%v", err)
	}

	return c.MqcProducer.Publish(topic, b)
}

func (c *component) Close() {
	//关闭所有公共组件
}
