package processor

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Sirupsen/logrus"
	"github.com/dchest/uniuri"
)

// MetadataQueueProcessor is embedding the Processor structure and it knows how to handle Metadata kafka messages.
type MetadataQueueProcessor struct {
	Processor *QueueProcessor
}

func NewMetadataQueueProcessor(queueAddress string, routingHeader string, topic string, group string, sourceConcurrentProcessing bool, combiner *MessageCombiner, forwarder *MessageForwarder) *MetadataQueueProcessor {
	p := NewQueueProcessor(queueAddress, routingHeader, topic, group, sourceConcurrentProcessing, combiner, forwarder)
	return &MetadataQueueProcessor{p}
}

func (mqp *MetadataQueueProcessor) ProcessMsg(m consumer.Message) {

	tid, err := extractTID(m.Headers)
	if err != nil {
		logrus.Infof("Couldn't extract transaction id: %s", err.Error())
		tid = "tid_" + uniuri.NewLen(10) + "_post_publication_combiner"
		logrus.Infof("Generated tid: %d", tid)
	}
	m.Headers["X-Request-Id"] = tid

	//decide based on the origin system header - whether you want to process the message or not
	o := m.Headers["Origin-System-Id"]
	if o != "http://cmdb.ft.com/systems/binding-service" {
		logrus.Printf("%v - Skipped unsupported annotations with Origin-System-Id: %v. ", tid, o)
		return
	}

	//parse message - collect data, then forward it to the next queue
	var ann model.Annotations
	b := []byte(m.Body)
	if err := json.Unmarshal(b, &ann); err != nil {
		logrus.Errorf("Could not unmarshall message with ID=%v, error=%v", m.Headers["Message-Id"], err.Error())
		return
	}

	//combine data
	combinedMSG, err := mqp.Processor.Combiner.enrichWithContent(ann)
	if err != nil {
		logrus.Errorf("%v - Error obtaining the combined message. Content couldn't get read. Message will be skipped. %v", tid, err)
		return
	}

	//forward data
	err = mqp.Processor.Forwarder.forwardMsg(m.Headers, &combinedMSG)
	if err != nil {
		logrus.Errorf("%v - Error sending transformed message to queue: %v", tid, err)
		return
	}
	logrus.Printf("%v - Mapped and sent for uuid: %v", tid, combinedMSG.UUID)
}
