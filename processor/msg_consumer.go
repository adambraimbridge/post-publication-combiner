package processor

import (
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"net/http"
)

type QConsumer interface {
	ProcessMsg(m consumer.Message)
}

type KafkaQConsumer struct {
	Consumer consumer.MessageConsumer
	dest     chan<- *KafkaQMessage
	msgType  string
}

type KafkaQMessage struct {
	msgType string
	msg     consumer.Message
}

func NewKafkaQConsumer(cConf consumer.QueueConfig, ch chan<- *KafkaQMessage, client *http.Client) *KafkaQConsumer {

	kc := KafkaQConsumer{msgType: cConf.Topic, dest: ch}
	kc.Consumer = consumer.NewConsumer(cConf, kc.ProcessMsg, client)
	return &kc
}

func (c *KafkaQConsumer) ProcessMsg(m consumer.Message) {
	c.dest <- &KafkaQMessage{msgType: c.msgType, msg: m}
}
