package nsq

import "github.com/nsqio/go-nsq"

type MqcHandler interface { //接入nsq
	HandleMessage(message *nsq.Message) error
}
