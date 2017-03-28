package processor

import (
	"encoding/json"
	"errors"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"net/http"
)

type Processor interface {
	ProcessMsg(m consumer.Message)
}

// QueueProcessor is a structure meant to process and forward messages to a kafka queue
type QueueProcessor struct {
	QConf       consumer.QueueConfig
	MsgProducer producer.MessageProducer
}

func NewQueueProcessor(cConf consumer.QueueConfig, pConf producer.MessageProducerConfig, c *http.Client) *QueueProcessor {
	p := producer.NewMessageProducerWithHTTPClient(pConf, c)
	return &QueueProcessor{cConf, p}
}

func extractTID(headers map[string]string) (string, error) {
	header := headers["X-Request-Id"]
	if header == "" {
		return "", errors.New("X-Request-Id header could not be found.")
	}
	return header, nil
}

func (p *QueueProcessor) forwardMsg(headers map[string]string, model *model.CombinedModel) error {
	// marshall message
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	// add special message type
	headers["Message-Type"] = "cms-combined-content-published"
	return p.MsgProducer.SendMessage(model.UUID, producer.Message{Headers: headers, Body: string(b)})
}

func NewQueueConsumerConfig(queueAddress string, routingHeader string, topic string, group string, sourceConcurrentProcessing bool) consumer.QueueConfig {
	return consumer.QueueConfig{
		Addrs:                []string{queueAddress},
		Group:                group,
		Topic:                topic,
		Queue:                routingHeader,
		ConcurrentProcessing: sourceConcurrentProcessing,
	}
}

func NewProducerConfig(proxyAddress string, topic string, routingHeader string) producer.MessageProducerConfig {
	return producer.MessageProducerConfig{
		Addr:  proxyAddress,
		Topic: topic,
		Queue: routingHeader,
	}
}
