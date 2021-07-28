package mqc

type MqcInvoker interface { //服务端要实现的Mqc调用程序组件
	Publish(topic string, body []byte) error
	Close()
}
