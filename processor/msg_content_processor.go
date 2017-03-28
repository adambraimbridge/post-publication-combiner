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
	"strings"
)

// ContentQueueProcessor is the implementation of the Processor interface and it knows how to communicate with Content kafka queue.
type ContentQueueProcessor struct {
	Processor             *QueueProcessor
	httpClient            *http.Client
	AnnotationsApiAddress utils.ApiURL
}

func NewContentQueueProcessor(cConf consumer.QueueConfig, pConf producer.MessageProducerConfig, annApiBaseURL string, annApiEndpoint string, client *http.Client) *ContentQueueProcessor {

	p := NewQueueProcessor(cConf, pConf, client)
	annAddress := utils.ApiURL{
		BaseURL:  annApiBaseURL,
		Endpoint: annApiEndpoint,
	}
	return &ContentQueueProcessor{p, client, annAddress}
}

func (cqp *ContentQueueProcessor) ProcessMsg(m consumer.Message) {

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
	if supportedType(cm.ContentURI) {
		//combine data
		combinedMSG, err := combiner.GetCombinedModelForContent(cqp.AnnotationsApiAddress, cqp.httpClient, cm.ContentModel)
		if err != nil {
			logrus.Errorf("%v - Error obtaining the combined message. Metadata could not be read. Message will be skipped. %v", tid, err)
			return
		}

		//forward data
		err = cqp.Processor.forwardMsg(m.Headers, &combinedMSG)
		if err != nil {
			logrus.Errorf("%v - Error sending transformed message to queue: %v", tid, err)
			return
		}
		logrus.Printf("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
		return
	}

	logrus.Printf("%v - Skipped unsupported content with contentUri: %v. ", tid, cm.ContentURI)
}

func supportedType(contentUri string) (supported bool) {

	if strings.Contains(contentUri, "methode-article-mapper") || strings.Contains(contentUri, "wordpress-article-mapper") || strings.Contains(contentUri, "brightcove-video-model-mapper") {
		return true
	}

	return false
}
