package processor

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/combiner"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/Sirupsen/logrus"
	"github.com/dchest/uniuri"
	"net/http"
)

// MetadataQueueProcessor is embedding the Processor structure and it knows how to handle Metadata kafka messages.
type MetadataQueueProcessor struct {
	*QueueProcessor
	httpClient            *http.Client
	DocStoreApiAddress    utils.ApiURL
	AnnotationsApiAddress utils.ApiURL
	SupportedHeaders      []string
}

func NewMetadataQueueProcessor(cConf consumer.QueueConfig, pConf producer.MessageProducerConfig, annApiBaseURL string, annApiEndpoint string, docStoreApiBaseURL string, docStoreApiEndpoint string, client *http.Client, supportedHeaders []string) *MetadataQueueProcessor {

	mqp := MetadataQueueProcessor{
		httpClient: client,
		DocStoreApiAddress: utils.ApiURL{
			BaseURL:  docStoreApiBaseURL,
			Endpoint: docStoreApiEndpoint,
		},
		AnnotationsApiAddress: utils.ApiURL{
			BaseURL:  annApiBaseURL,
			Endpoint: annApiEndpoint,
		},
		SupportedHeaders: supportedHeaders,
	}

	p := NewQueueProcessor(cConf, mqp.processMsg, pConf, client)
	mqp.QueueProcessor = p

	return &mqp
}

func (mqp *MetadataQueueProcessor) processMsg(m consumer.Message) {

	tid, err := extractTID(m.Headers)
	if err != nil {
		logrus.Infof("Couldn't extract transaction id: %s", err.Error())
		tid = "tid_" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}
	m.Headers["X-Request-Id"] = tid
	h := m.Headers["Origin-System-Id"]

	//decide based on the origin system header - whether you want to process the message or not
	if headerIsSupported(h, mqp.SupportedHeaders) {
		//parse message - collect data, then forward it to the next queue
		var ann model.Annotations
		b := []byte(m.Body)
		if err := json.Unmarshal(b, &ann); err != nil {
			logrus.Errorf("Could not unmarshall message with ID=%v, error=%v", m.Headers["Message-Id"], err.Error())
			return
		}

		//combine data
		combinedMSG, err := combiner.GetCombinedModelForAnnotations(mqp.DocStoreApiAddress, mqp.AnnotationsApiAddress, mqp.httpClient, ann)
		if err != nil {
			logrus.Errorf("%v - Error obtaining the combined message. Content couldn't get read. Message will be skipped. %v", tid, err)
			return
		}

		//forward data
		err = mqp.forwardMsg(m.Headers, &combinedMSG)
		if err != nil {
			logrus.Errorf("%v - Error sending transformed message to queue: %v", tid, err)
			return
		}
		logrus.Printf("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
		return
	}

	logrus.Printf("%v - Skipped unsupported annotations with Origin-System-Id: %v. ", tid, h)
}

func headerIsSupported(header string, supportedHeaders []string) (supported bool) {

	for _, h := range supportedHeaders {
		if h == header {
			return true
		}
	}

	return false
}
