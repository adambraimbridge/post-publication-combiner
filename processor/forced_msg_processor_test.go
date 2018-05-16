package processor

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"fmt"
	"testing"
	testLogger "github.com/Financial-Times/go-logger/test"
	"github.com/stretchr/testify/assert"
	"errors"
)

func TestForceMessageWithTID(t *testing.T) {
	allowedContentTypes := []string{"Article", "Video"}

	dummyDataCombiner := DummyDataCombiner{
		data: CombinedModel{
			UUID:    "some_uuid",
			Content: ContentModel{"uuid": "some_uuid", "title": "simple title", "type": "Article"},
			Metadata: []Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					},
				},
			},
		}}
	tid := "transaction_id_1"
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": CombinerMessageType, "X-Request-Id": tid, "Origin-System-Id": CombinerOrigin, "Content-Type": ContentType},
		Body:    `{"uuid":"some_uuid","contentUri":"","lastModified":"","markedDeleted":"","content":{"uuid":"some_uuid","title":"simple title","type":"Article"},"metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995"}}]}`,
	}

	dummyMsgProducer := DummyMsgProducer{t: t, expUUID: dummyDataCombiner.data.UUID, expTID: tid, expMsg: expMsg}
	p := &ForcedMsgProcessor{DataCombiner: dummyDataCombiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("dummyDataCombiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(dummyDataCombiner.data.UUID, tid)
	assert.NoError(t, err)

	assert.Equal(t, "info", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", expMsg.Headers["X-Request-Id"], dummyDataCombiner.data.UUID))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestForceMessageWithoutTID(t *testing.T) {
	allowedContentTypes := []string{"Article", "Video"}

	dummyDataCombiner := DummyDataCombiner{
		data: CombinedModel{
			UUID:    "some_uuid",
			Content: ContentModel{"uuid": "some_uuid", "title": "simple title", "type": "Article"},
			Metadata: []Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					},
				},
			},
		}}

	emptyTID := ""
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": CombinerMessageType, "X-Request-Id": "[ignore]", "Origin-System-Id": CombinerOrigin, "Content-Type": ContentType},
		Body:    `{"uuid":"some_uuid","contentUri":"","lastModified":"","markedDeleted":"","content":{"uuid":"some_uuid","title":"simple title","type":"Article"},"metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995"}}]}`,
	}

	dummyMsgProducer := DummyMsgProducer{t: t, expUUID: dummyDataCombiner.data.UUID, expMsg: expMsg}
	p := &ForcedMsgProcessor{DataCombiner: dummyDataCombiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("dummyDataCombiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(dummyDataCombiner.data.UUID, emptyTID)
	assert.NoError(t, err)

	assert.Equal(t, "info", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", expMsg.Headers["X-Request-Id"], dummyDataCombiner.data.UUID))
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessageCombinerError(t *testing.T) {
	allowedContentTypes := []string{"Article", "Video"}
	combiner := DummyDataCombiner{err: errors.New("some error")}
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": CombinerMessageType, "X-Request-Id": "[ignore]", "Origin-System-Id": CombinerOrigin, "Content-Type": ContentType},
		Body:    `{"uuid":"some_uuid","contentUri":"","lastModified":"","markedDeleted":"","content":{"uuid":"some_uuid","title":"simple title","type":"Article"},"metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995"}}]}`,
	}
	dummyMsgProducer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &ForcedMsgProcessor{DataCombiner: combiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("combiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(combiner.data.UUID, "")
	assert.Equal(t, combiner.err, err)

	assert.Equal(t, "error", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, "Error obtaining the combined message, it will be skipped.")
	assert.Equal(t, hook.LastEntry().Data["error"].(error).Error(), "some error")
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessageNotFoundError(t *testing.T) {

	allowedContentTypes := []string{"Article", "Video"}
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": CombinerMessageType, "X-Request-Id": "[ignore]", "Origin-System-Id": CombinerOrigin, "Content-Type": ContentType},
		Body:    `{"uuid":"some_uuid","contentUri":"","lastModified":"","markedDeleted":"","content":{"uuid":"some_uuid","title":"simple title","type":"Article"},"metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995"}}]}`,
	}
	combiner := DummyDataCombiner{}
	dummyMsgProducer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &ForcedMsgProcessor{DataCombiner: combiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("combiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	uuid := "80fb3e57-8d3b-4f07-bbb6-8788452d63cb"
	err := p.ForceMessagePublish(uuid, "")
	assert.Equal(t, NotFoundError, err)

	assert.Equal(t, "error", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("Could not find content with uuid %s.", uuid))
	assert.Equal(t, hook.LastEntry().Data["error"].(error).Error(), "Content not found")
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessage_FilteringError(t *testing.T) {
	allowedContentTypes := []string{"Article", "Video"}
	combiner := DummyDataCombiner{
		data: CombinedModel{
			UUID: "80fb3e57-8d3b-4f07-bbb6-8788452d63cb",
			// Content Placeholders - marked with Content type - shouldn't get into the Combined queue
			Content: ContentModel{"uuid": "80fb3e57-8d3b-4f07-bbb6-8788452d63cb", "title": "simple title", "type": "Content"},
			Metadata: []Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					},
				},
			},
		}}
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": CombinerMessageType, "X-Request-Id": "[ignore]", "Origin-System-Id": CombinerOrigin, "Content-Type": ContentType},
		Body:    `{"uuid":"some_uuid","contentUri":"","lastModified":"","markedDeleted":"","content":{"uuid":"some_uuid","title":"simple title","type":"Article"},"metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995"}}]}`,
	}
	dummyMsgProducer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &ForcedMsgProcessor{DataCombiner: combiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("combiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	uuid := "80fb3e57-8d3b-4f07-bbb6-8788452d63cb"
	err := p.ForceMessagePublish(uuid, "")
	assert.Equal(t, InvalidContentTypeError, err)

	assert.Equal(t, "info", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, "Skipped unsupported content with type: Content")
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessageProducerError(t *testing.T) {
	allowedContentTypes := []string{"Article", "Video"}

	dummyDataCombiner := DummyDataCombiner{
		data: CombinedModel{
			UUID:    "some_uuid",
			Content: ContentModel{"uuid": "some_uuid", "title": "simple title", "type": "Article"},
			Metadata: []Annotation{
				{
					Thing: Thing{
						ID:        "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
						PrefLabel: "Barclays",
						Types: []string{"http://base-url/core/Thing",
							"http://base-url/concept/Concept",
							"http://base-url/organisation/Organisation",
							"http://base-url/company/Company",
							"http://base-url/company/PublicCompany",
						},
						Predicate: "http://base-url/about",
						ApiUrl:    "http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995",
					},
				},
			},
		}}
	dummyMsgProducer := DummyMsgProducer{t: t, expError: errors.New("some error")}
	p := &ForcedMsgProcessor{DataCombiner: dummyDataCombiner, Processor: NewForwarder(dummyMsgProducer, allowedContentTypes)}

	hook := testLogger.NewTestHook("dummyDataCombiner")
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(dummyDataCombiner.data.UUID, "")
	assert.Equal(t, dummyMsgProducer.expError, err)

	assert.Equal(t, "error", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, "Error sending transformed message to queue.")
	assert.Equal(t, hook.LastEntry().Data["error"].(error).Error(), "some error")
	assert.Equal(t, 2, len(hook.Entries))
}
