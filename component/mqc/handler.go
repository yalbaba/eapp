package mqc

import "github.com/nsqio/go-nsq"

type MqcHandler interface {
	HandleMessage(msg *nsq.Message) error
}
