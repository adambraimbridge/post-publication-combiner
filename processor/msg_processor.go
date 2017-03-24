package processor

import (
	"errors"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

type Processor interface {
	ProcessMsg(m consumer.Message)
}

// QueueProcessor is a structure meant to receive, process and forward messages from a kafka queue
type QueueProcessor struct {
	QConf     consumer.QueueConfig
	Combiner  *MessageCombiner
	Forwarder *MessageForwarder
}

func NewQueueProcessor(queueAddress string, routingHeader string, topic string, group string, sourceConcurrentProcessing bool, combiner *MessageCombiner, forwarder *MessageForwarder) *QueueProcessor {
	c := newQueueConsumerConfig(queueAddress, routingHeader, topic, group, sourceConcurrentProcessing)
	return &QueueProcessor{c, combiner, forwarder}
}

func extractTID(headers map[string]string) (string, error) {
	header := headers["X-Request-Id"]
	if header == "" {
		return "", errors.New("X-Request-Id header could not be found.")
	}
	return header, nil
}

func newQueueConsumerConfig(queueAddress string, routingHeader string, topic string, group string, sourceConcurrentProcessing bool) consumer.QueueConfig {
	return consumer.QueueConfig{
		Addrs:                []string{queueAddress},
		Group:                group,
		Topic:                topic,
		Queue:                routingHeader,
		ConcurrentProcessing: sourceConcurrentProcessing,
	}
}
