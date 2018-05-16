package processor

import (
	"github.com/Financial-Times/go-logger"
	"github.com/dchest/uniuri"
	"github.com/Financial-Times/message-queue-go-producer/producer"
)

const (
	CombinerOrigin = "forced-combined-msg"
	ContentType    = "application/json"
)

type ForcedMsgProcessorI interface {
	ForceMessagePublish(uuid string, tid string) error
}

type ForcedMsgProcessor struct {
	DataCombiner DataCombinerI
	Processor    Forwarder
}

func NewForcedMsgProcessor(dataCombiner DataCombinerI, producer producer.MessageProducer, whitelistedContentTypes []string) *ForcedMsgProcessor {
	return &ForcedMsgProcessor{DataCombiner: dataCombiner, Processor: NewForwarder(producer, whitelistedContentTypes)}
}

func (p *ForcedMsgProcessor) ForceMessagePublish(uuid string, tid string) error {

	if tid == "" {
		tid = "tid_force_publish" + uniuri.NewLen(10) + "_post_publication_combiner"
		logger.WithTransactionID(tid).WithUUID(uuid).Infof("Generated tid: %s", tid)
	}

	h := map[string]string{
		"X-Request-Id":     tid,
		"Content-Type":     ContentType,
		"Origin-System-Id": CombinerOrigin,
	}

	//get combined message
	combinedMSG, err := p.DataCombiner.GetCombinedModel(uuid)
	if err != nil {
		logger.WithTransactionID(tid).WithUUID(uuid).WithError(err).Errorf("%v - Error obtaining the combined message, it will be skipped.", tid)
		return err
	}

	if combinedMSG.Content.getUUID() == "" && combinedMSG.Metadata == nil {
		err := NotFoundError
		logger.WithTransactionID(tid).WithUUID(uuid).WithError(err).Errorf("%v - Could not find content with uuid %s.", tid, uuid)
		return err
	}

	//forward data
	return p.Processor.filterAndForwardMsg(h, &combinedMSG, tid)
}
