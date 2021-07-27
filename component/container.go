package component

import (
	"eapp/component/mqc"
	"eapp/component/rpc"
	"eapp/global_config"
	"eapp/logger"
	"fmt"
	jsoniter "github.com/json-iterator/go"
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

//-----------自定义的框架容器实现-----------
type component struct {
	RpcInvoker  rpc.RpcInvoker //RPC组件（可以集成任何mq中中间件）
	MqcProducer mqc.MqcInvoker //MQC组件（可以集成任何mq中中间件）
	Num         int

	logger.ILogger
}

func NewComponent(l logger.ILogger) IComponent {
	c := &component{}
	c.ILogger = l
	return c
}

func (c *component) NewMqcProducer() {
	c.MqcProducer = mqc.NewNsqInvoker(global_config.Conf.MqcService.Host)
}

func (c *component) NewRpcInvoker() {
	c.RpcInvoker = rpc.NewRpcInvoker(global_config.Conf)
}

func (c *component) RpcRequest(cluster, service string, header map[string]string, input map[string]interface{}, failFast bool) (interface{}, error) {
	return c.RpcInvoker.Request(cluster, service, header, input, failFast)
}

func (c *component) Send(topic string, body map[string]interface{}) error {
	b, err := jsoniter.Marshal(body)
	if err != nil {
		return fmt.Errorf("消息转json失败,err:%v", err)
	}

	return c.MqcProducer.Publish(topic, b)
}

func (c *component) Close() {
	//关闭所有公共组件
	if c.MqcProducer != nil {
		c.MqcProducer.Close()
	}
	if c.RpcInvoker != nil {
		c.RpcInvoker.Close()
	}
}
