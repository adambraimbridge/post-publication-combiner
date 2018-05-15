package processor

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/go-logger"
	"encoding/json"
)

type Processor struct {
	MsgProducer           producer.MessageProducer
	SupportedContentTypes []string
}

func NewProcessor(msgProducer producer.MessageProducer, supportedContentTypes []string) Processor {
	return Processor{
		MsgProducer:           msgProducer,
		SupportedContentTypes: supportedContentTypes,
	}
}


func (p *Processor) filterAndForwardMsg(headers map[string]string, combinedMSG *CombinedModel, tid string) error {

	if combinedMSG.Content != nil && !isTypeAllowed(p.SupportedContentTypes, combinedMSG.Content.getType()) {
		logger.WithTransactionID(tid).Infof("%v - Skipped unsupported content with type: %v", tid, combinedMSG.Content.getType())
		return InvalidContentTypeError
	}

	//forward data
	err := p.forwardMsg(headers, combinedMSG)
	if err != nil {
		logger.WithTransactionID(tid).WithError(err).Errorf("%v - Error sending transformed message to queue.", tid)
		return err
	}
	logger.WithTransactionID(tid).Infof("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
	return nil
}

func isTypeAllowed(allowedTypes []string, value string) bool {
	return contains(allowedTypes, value)
}

func (p *Processor) forwardMsg(headers map[string]string, model *CombinedModel) error {
	// marshall message
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	// add special message type
	headers["Message-Type"] = CombinerMessageType
	return p.MsgProducer.SendMessage(model.UUID, producer.Message{Headers: headers, Body: string(b)})
}

func contains(array []string, element string) bool {
	for _, e := range array {
		if element == e {
			return true
		}
	}
	return false
}
