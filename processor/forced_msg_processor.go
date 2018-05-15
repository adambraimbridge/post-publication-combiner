package processor

import (
	"net/http"
	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/dchest/uniuri"
	"github.com/Financial-Times/message-queue-go-producer/producer"
)

type ForcedMsgProcessorI interface {
	ForceMessagePublish(uuid string, tid string) error
}

type ForcedMsgProcessor struct {
	client       *http.Client
	DataCombiner DataCombinerI
	Processor    Processor
}

func NewForcedMsgProcessor(prodConf producer.MessageProducerConfig, docStoreApiUrl utils.ApiURL, annApiUrl utils.ApiURL, whitelistedContentTypes []string, c *http.Client) *ForcedMsgProcessor {
	p := producer.NewMessageProducerWithHTTPClient(prodConf, c)

	var cRetriever contentRetrieverI = dataRetriever{docStoreApiUrl, c}
	var mRetriever metadataRetrieverI = dataRetriever{annApiUrl, c}

	var combiner DataCombinerI = DataCombiner{
		ContentRetriever:  cRetriever,
		MetadataRetriever: mRetriever,
	}
	return &ForcedMsgProcessor{client: c, DataCombiner: combiner, Processor: NewProcessor(p, whitelistedContentTypes)}
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
