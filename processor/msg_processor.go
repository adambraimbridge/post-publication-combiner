package processor

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Sirupsen/logrus"
	"github.com/dchest/uniuri"
	"net/http"
)

const (
	CombinerMessageType = "cms-combined-content-published"
)

type Processor interface {
	ProcessMsg(m consumer.Message)
}

// QueueProcessor is a structure meant to process and forward messages to a kafka queue
type QueueProcessor struct {
	MessageConsumer consumer.MessageConsumer
	MsgProducer     producer.MessageProducer
}

func NewQueueProcessor(cConf consumer.QueueConfig, handleFunc func(m consumer.Message), pConf producer.MessageProducerConfig, cl *http.Client) *QueueProcessor {

	p := producer.NewMessageProducerWithHTTPClient(pConf, cl)
	c := consumer.NewConsumer(cConf, handleFunc, cl)

	return &QueueProcessor{c, p}
}

func NewQueueConsumerConfig(queueAddress string, routingHeader string, topic string, group string, sourceConcurrentProcessing bool) consumer.QueueConfig {
	return consumer.QueueConfig{
		Addrs: []string{queueAddress},
		Group: group,
		Topic: topic,
		Queue: routingHeader,
	}
}

func NewProducerConfig(proxyAddress string, topic string, routingHeader string) producer.MessageProducerConfig {
	return producer.MessageProducerConfig{
		Addr:  proxyAddress,
		Topic: topic,
		Queue: routingHeader,
	}
}

func extractTID(headers map[string]string) string {
	tid := headers["X-Request-Id"]

	if tid == "" {
		logrus.Infof("Couldn't extract transaction id - X-Request-Id header could not be found.")
		tid = "tid_" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}

	return tid
}

func contains(array []string, element string) bool {

	for _, e := range array {
		if e == element {
			return true
		}
	}

	return false
}

func (p *QueueProcessor) forwardMsg(headers map[string]string, model *model.CombinedModel) error {
	// marshall message
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	// add special message type
	headers["Message-Type"] = CombinerMessageType
	return p.MsgProducer.SendMessage(model.UUID, producer.Message{Headers: headers, Body: string(b)})
}
