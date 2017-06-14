package processor

import (
	"encoding/json"
	"errors"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/Sirupsen/logrus"
	"github.com/dchest/uniuri"
	"net/http"
	"reflect"
	"strings"
)

const (
	CombinerMessageType = "cms-combined-content-published"
	// PlatformV1 V1 platform (Falcon)
	PlatformV1 = "v1"

	// PlatformVideo current video platform
	PlatformVideo = "next-video"
)

// NotFoundError used when the content can not be found by the platform
var NotFoundError = errors.New("Content not found")
var InvalidContentTypeError = errors.New("Invalid content type")

type Processor interface {
	ProcessMessages()
	ForceMessagePublish(uuid, transactionID, platformVersion string) error
}

type MsgProcessor struct {
	src          <-chan *KafkaQMessage
	client       *http.Client
	config       MsgProcessorConfig
	DataCombiner DataCombinerI
	MsgProducer  producer.MessageProducer
}

type MsgProcessorConfig struct {
	SupportedContentTypes []string
	SupportedContentURIs  []string
	SupportedHeaders      []string
	ContentTopic          string
	MetadataTopic         string
}

func NewMsgProcessorConfig(supportedContentTypes []string, supportedURIs []string, supportedHeaders []string, contentTopic string, metadataTopic string) MsgProcessorConfig {
	return MsgProcessorConfig{
		SupportedContentTypes: supportedContentTypes,
		SupportedContentURIs:  supportedURIs,
		SupportedHeaders:      supportedHeaders,
		ContentTopic:          contentTopic,
		MetadataTopic:         metadataTopic,
	}
}

func NewMsgProcessor(prodConf producer.MessageProducerConfig, srcCh <-chan *KafkaQMessage, docStoreApiUrl utils.ApiURL, annApiUrl utils.ApiURL, c *http.Client, config MsgProcessorConfig) *MsgProcessor {
	p := producer.NewMessageProducerWithHTTPClient(prodConf, c)

	var cRetriever contentRetrieverI = dataRetriever{docStoreApiUrl, c}
	var mRetriever metadataRetrieverI = dataRetriever{annApiUrl, c}

	var combiner DataCombinerI = DataCombiner{
		ContentRetriever:  cRetriever,
		MetadataRetriever: mRetriever,
	}
	return &MsgProcessor{src: srcCh, MsgProducer: p, config: config, client: c, DataCombiner: combiner}
}

func NewProducerConfig(proxyAddress string, topic string, routingHeader string) producer.MessageProducerConfig {
	return producer.MessageProducerConfig{
		Addr:  proxyAddress,
		Topic: topic,
		Queue: routingHeader,
	}
}

func (p *MsgProcessor) ProcessMessages() {
	for {
		m := <-p.src
		if m.msgType == p.config.ContentTopic {
			p.processContentMsg(m.msg)
		} else if m.msgType == p.config.MetadataTopic {
			p.processMetadataMsg(m.msg)
		}
	}
}

func (p *MsgProcessor) ForceMessagePublish(uuid, tid, platformVersion string) error {

	if tid == "" {
		tid = "tid_force_publish" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}

	h := map[string]string{
		"X-Request-Id":     tid,
		"Origin-System-Id": "force-publish",
	}

	//get combined message
	combinedMSG, err := p.DataCombiner.GetCombinedModel(uuid, platformVersion)
	if err != nil {
		logrus.Errorf("%v - Error obtaining the combined message, it will be skipped. %v", tid, err)
		return err
	}

	if combinedMSG.Content.UUID == "" && combinedMSG.Metadata == nil {
		err := NotFoundError
		logrus.Errorf("%v - Could not find content with uuid %s. %v", tid, uuid, err)
		return err
	}

	//forward data
	return p.filterAndForwardMsg(h, &combinedMSG, tid)
}

func (p *MsgProcessor) processContentMsg(m consumer.Message) {

	tid := extractTID(m.Headers)
	m.Headers["X-Request-Id"] = tid

	//parse message - collect data, then forward it to the next queue
	var cm model.MessageContent
	b := []byte(m.Body)
	if err := json.Unmarshal(b, &cm); err != nil {
		logrus.Errorf("Could not unmarshall message with TID=%v, error=%v", tid, err.Error())
		return
	}

	// wordpress, next-video, methode-article - the system origin is not enough to help us filtering. Filter by contentUri.
	if !containsSubstringOf(p.config.SupportedContentURIs, cm.ContentURI) {
		logrus.Infof("%v - Skipped unsupported content with contentUri: %v. ", tid, cm.ContentURI)
		return
	}

	//handle delete events
	if reflect.DeepEqual(cm.ContentModel, model.ContentModel{}) {
		sl := strings.Split(cm.ContentURI, "/")
		cm.ContentModel.UUID = sl[len(sl)-1]
		cm.ContentModel.MarkedDeleted = true
	}

	if cm.ContentModel.UUID == "" {
		logrus.Errorf("UUID not found after message marshalling, skipping message with TID=%v.", tid)
		return
	}

	//combine data
	combinedMSG, err := p.DataCombiner.GetCombinedModelForContent(cm.ContentModel, getPlatformVersion(cm.ContentURI))
	if err != nil {
		logrus.Errorf("%v - Error obtaining the combined message. Metadata could not be read. Message will be skipped. %v", tid, err)
		return
	}

	//forward data
	p.filterAndForwardMsg(m.Headers, &combinedMSG, tid)
}

func (p *MsgProcessor) processMetadataMsg(m consumer.Message) {

	tid := extractTID(m.Headers)
	m.Headers["X-Request-Id"] = tid
	h := m.Headers["Origin-System-Id"]

	//decide based on the origin system header - whether you want to process the message or not
	if !containsSubstringOf(p.config.SupportedHeaders, h) {
		logrus.Infof("%v - Skipped unsupported annotations with Origin-System-Id: %v. ", tid, h)
		return
	}

	//parse message - collect data, then forward it to the next queue
	var ann model.Annotations
	b := []byte(m.Body)
	if err := json.Unmarshal(b, &ann); err != nil {
		logrus.Errorf("Could not unmarshall message with TID=%v, error=%v", tid, err.Error())
		return
	}

	//combine data
	combinedMSG, err := p.DataCombiner.GetCombinedModelForAnnotations(ann, getPlatformVersion(h))
	if err != nil {
		logrus.Errorf("%v - Error obtaining the combined message. Content couldn't get read. Message will be skipped. %v", tid, err)
		return
	}
	p.filterAndForwardMsg(m.Headers, &combinedMSG, tid)
}

func (p *MsgProcessor) filterAndForwardMsg(headers map[string]string, combinedMSG *model.CombinedModel, tid string) error {
	if !combinedMSG.Content.MarkedDeleted && !isTypeAllowed(p.config.SupportedContentTypes, combinedMSG.Content.Type) {
		logrus.Infof("%v - Skipped unsupported content with type: %v", tid, combinedMSG.Content.Type)
		return InvalidContentTypeError
	}

	//forward data
	err := p.forwardMsg(headers, combinedMSG)
	if err != nil {
		logrus.Errorf("%v - Error sending transformed message to queue: %v", tid, err)
		return err
	}
	logrus.Infof("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
	return nil
}

func isTypeAllowed(allowedTypes []string, value string) bool {
	return contains(allowedTypes, value)
}

func (p *MsgProcessor) forwardMsg(headers map[string]string, model *model.CombinedModel) error {
	// marshall message
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	// add special message type
	headers["Message-Type"] = CombinerMessageType
	return p.MsgProducer.SendMessage(model.UUID, producer.Message{Headers: headers, Body: string(b)})
}

func extractTID(headers map[string]string) string {
	tid := headers["X-Request-Id"]

	if tid == "" {
		logrus.Infof("Couldn't extract transaction id - X-Request-Id header could not be found.")
		tid = "tid_" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}

	return tid
}

func containsSubstringOf(array []string, element string) bool {
	for _, e := range array {
		if strings.Contains(element, e) {
			return true
		}
	}
	return false
}

func contains(array []string, element string) bool {
	for _, e := range array {
		if element == e {
			return true
		}
	}
	return false
}

func getPlatformVersion(str string) string {
	if strings.Contains(str, "video") {
		return PlatformVideo
	}
	return PlatformV1
}
