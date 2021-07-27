package mqc

import "github.com/nsqio/go-nsq"

type MqcHandler interface { //客户端需要实现的处理接口（本框架实现接入nsq）
	HandleMessage(message *nsq.Message) error
}
