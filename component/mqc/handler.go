package mqc

import "github.com/nsqio/go-nsq"

type MqcHandler interface { //客户端需要实现的处理接口
	Receive(msg *nsq.Message) error
}
