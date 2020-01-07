package processor

import (
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	logger "github.com/Financial-Times/go-logger/v2"
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
	logConf := logger.KeyNamesConfig{KeyTime: "@time"}
	l := logger.NewUPPLogger("post-publication-combiner", "INFO", logConf)

	kc := KafkaQConsumer{msgType: cConf.Topic, dest: ch}
	kc.Consumer = consumer.NewConsumer(cConf, kc.ProcessMsg, client, l)
	return &kc
}

func (c *KafkaQConsumer) ProcessMsg(m consumer.Message) {
	c.dest <- &KafkaQMessage{msgType: c.msgType, msg: m}
}
