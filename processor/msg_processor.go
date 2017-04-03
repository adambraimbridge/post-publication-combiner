package processor

import (
	"encoding/json"
	"errors"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Sirupsen/logrus"
	"net/http"
"sync"
"os"
"os/signal"
"syscall"
)

type Processor interface {
	ProcessMessages()
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

func (qp *QueueProcessor) ProcessMessages() {

	if qp == nil {
		logrus.Errorf("Consumer is not created. Messages won't be processed.")
		return
	}

	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)

	go func() {
		qp.MessageConsumer.Start()
		consumerWaitGroup.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	qp.MessageConsumer.Stop()
	consumerWaitGroup.Wait()

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
