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
	"reflect"
	"strings"
)

// ContentQueueProcessor is the implementation of the Processor interface and it knows how to communicate with Content kafka queue.
type ContentQueueProcessor struct {
	*QueueProcessor
	httpClient            *http.Client
	AnnotationsApiAddress utils.ApiURL
	SupportedContentURIs  []string
}

func NewContentQueueProcessor(cConf consumer.QueueConfig, pConf producer.MessageProducerConfig, annApiBaseURL string, annApiEndpoint string, client *http.Client, supportedContentURIs []string) *ContentQueueProcessor {

	cqp := ContentQueueProcessor{
		httpClient: client,
		AnnotationsApiAddress: utils.ApiURL{
			BaseURL:  annApiBaseURL,
			Endpoint: annApiEndpoint,
		},
		SupportedContentURIs: supportedContentURIs,
	}

	p := NewQueueProcessor(cConf, cqp.processMsg, pConf, client)
	cqp.QueueProcessor = p

	return &cqp
}

func (cqp *ContentQueueProcessor) processMsg(m consumer.Message) {

	tid, err := extractTID(m.Headers)
	if err != nil {
		logrus.Infof("Couldn't extract transaction id: %s", err.Error())
		tid = "tid_" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}
	m.Headers["X-Request-Id"] = tid

	//parse message - collect data, then forward it to the next queue
	var cm model.MessageContent
	b := []byte(m.Body)
	if err := json.Unmarshal(b, &cm); err != nil {
		logrus.Errorf("Could not unmarshall message with ID=%v, error=%v", m.Headers["Message-Id"], err.Error())
		return
	}

	// wordpress, brightcove, methode-article - the system origin is not enough to help us filtering. Filter by contentUri.
	if contentTypeIsSupported(cm.ContentURI, cqp.SupportedContentURIs) {

		//handle delete events
		if reflect.DeepEqual(cm.ContentModel, model.ContentModel{}) {
			sl := strings.Split(cm.ContentURI, "/")
			cm.ContentModel.UUID = sl[len(sl)-1]
			cm.ContentModel.MarkedDeleted = true
			return
		}

		if cm.ContentModel.UUID == "" {
			logrus.Errorf("UUID not found after message marshalling, skipping message with ID=%v.", m.Headers["Message-Id"])
			return
		}

		//combine data
		combinedMSG, err := combiner.GetCombinedModelForContent(cqp.AnnotationsApiAddress, cqp.httpClient, cm.ContentModel)
		if err != nil {
			logrus.Errorf("%v - Error obtaining the combined message. Metadata could not be read. Message will be skipped. %v", tid, err)
			return
		}

		//forward data
		err = cqp.forwardMsg(m.Headers, &combinedMSG)
		if err != nil {
			logrus.Errorf("%v - Error sending transformed message to queue: %v", tid, err)
			return
		}
		logrus.Printf("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
		return
	}

	logrus.Printf("%v - Skipped unsupported content with contentUri: %v. ", tid, cm.ContentURI)
}

func contentTypeIsSupported(contentUri string, supportedContentURIs []string) (supported bool) {

	for _, uri := range supportedContentURIs {
		if strings.Contains(contentUri, uri) {
			return true
		}
	}

	return false
}
