package processor

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"net/http"
)

type MessageForwarder struct {
	producer *producer.MessageProducer
}

func NewMessageForwarder(proxyAddress string, routingHeader string, topic string, httpClient *http.Client) *MessageForwarder {
	pConf := producer.MessageProducerConfig{
		//proxy address
		Addr:  proxyAddress,
		Topic: topic,
		Queue: routingHeader,
	}
	p := producer.NewMessageProducerWithHTTPClient(pConf, httpClient)
	return &MessageForwarder{&p}
}

func (f *MessageForwarder) forwardMsg(headers map[string]string, model *model.CombinedModel) error {
	// marshall message
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	// add special message type
	headers["Message-Type"] = "cms-combined-content-published"
	return (*f.producer).SendMessage(model.UUID, producer.Message{Headers: headers, Body: string(b)})
}
