package processor

import (
	"encoding/json"
	"errors"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/Sirupsen/logrus"
	testLogger "github.com/Sirupsen/logrus/hooks/test"
	"github.com/golang/go/src/pkg/fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessContentMsg_Unmarshall_Error(t *testing.T) {

	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `body`,
	}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p := &MsgProcessor{}
	p.processContentMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "Could not unmarshall message with TID=")
	assert.Equal(t, 1, len(hook.Entries))

}

func TestProcessContentMsg_UnSupportedContent(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"ididn’tdoanything","alternativeTitles":{"promotionalTitle":null},"type":null,"byline":"","brands":[{"id":"http://base-url/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}],"identifiers":[{"authority":"METHODE-authority","identifierValue":"41358e4b-6d05-4f44-9eaf-f6a542154110"}],"publishedDate":"2017-03-30T13:08:53.000Z","standfirst":null,"body":"<body><p>lorem ipsum<\/p>\n<\/body>","description":null,"mediaType":null,"pixelWidth":null,"pixelHeight":null,"internalBinaryUrl":null,"externalBinaryUrl":null,"members":null,"mainImage":null,"storyPackage":null,"contentPackage":null,"standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"comments":{"enabled":true},"copyright":null,"webUrl":null,"publishReference":"tid_id_1","lastModified":"2017-03-30T13:09:06.480Z","canBeSyndicated":"verify","firstPublishedDate":"2017-03-30T13:08:53.000Z","accessLevel":"subscribed","canBeDistributed":"yes"},"contentUri":"http://unsupported-content-uri/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	p := &MsgProcessor{config: config}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Skipped unsupported content with contentUri: %v.", m.Headers["X-Request-Id"], "http://unsupported-content-uri/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b"))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessContentMsg_SupportedContent_EmptyUUID(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":{"title":"ididn’tdoanything","alternativeTitles":{"promotionalTitle":null},"type":null,"byline":"","brands":[{"id":"http://base-url/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}],"identifiers":[{"authority":"METHODE-authority","identifierValue":"41358e4b-6d05-4f44-9eaf-f6a542154110"}],"publishedDate":"2017-03-30T13:08:53.000Z","standfirst":null,"body":"<body><p>lorem ipsum<\/p>\n<\/body>","description":null,"mediaType":null,"pixelWidth":null,"pixelHeight":null,"internalBinaryUrl":null,"externalBinaryUrl":null,"members":null,"mainImage":null,"storyPackage":null,"contentPackage":null,"standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"comments":{"enabled":true},"copyright":null,"webUrl":null,"publishReference":"tid_id_1","lastModified":"2017-03-30T13:09:06.480Z","canBeSyndicated":"verify","firstPublishedDate":"2017-03-30T13:08:53.000Z","accessLevel":"subscribed","canBeDistributed":"yes"},"contentUri":"http://wordpress-article-mapper/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	p := &MsgProcessor{config: config}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("UUID not found after message marshalling, skipping message with TID=%v.", m.Headers["X-Request-Id"]))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessContentMsg_Combiner_Errors(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"ididn’tdoanything","alternativeTitles":{"promotionalTitle":null},"type":null,"byline":"","brands":[{"id":"http://base-url/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}],"identifiers":[{"authority":"METHODE-authority","identifierValue":"41358e4b-6d05-4f44-9eaf-f6a542154110"}],"publishedDate":"2017-03-30T13:08:53.000Z","standfirst":null,"body":"<body><p>lorem ipsum<\/p>\n<\/body>","description":null,"mediaType":null,"pixelWidth":null,"pixelHeight":null,"internalBinaryUrl":null,"externalBinaryUrl":null,"members":null,"mainImage":null,"storyPackage":null,"contentPackage":null,"standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"comments":{"enabled":true},"copyright":null,"webUrl":null,"publishReference":"tid_id_1","lastModified":"2017-03-30T13:09:06.480Z","canBeSyndicated":"verify","firstPublishedDate":"2017-03-30T13:08:53.000Z","accessLevel":"subscribed","canBeDistributed":"yes"},"contentUri":"http://wordpress-article-mapper/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	combiner := DummyDataCombiner{err: errors.New("some error")}
	p := &MsgProcessor{config: config, DataCombiner: combiner}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Error obtaining the combined message. Metadata could not be read. Message will be skipped. %v", m.Headers["X-Request-Id"], combiner.err.Error()))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessContentMsg_Forwarder_Errors(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"ididn’tdoanything","alternativeTitles":{"promotionalTitle":null},"type":null,"byline":"","brands":[{"id":"http://base-url/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}],"identifiers":[{"authority":"METHODE-authority","identifierValue":"41358e4b-6d05-4f44-9eaf-f6a542154110"}],"publishedDate":"2017-03-30T13:08:53.000Z","standfirst":null,"body":"<body><p>lorem ipsum<\/p>\n<\/body>","description":null,"mediaType":null,"pixelWidth":null,"pixelHeight":null,"internalBinaryUrl":null,"externalBinaryUrl":null,"members":null,"mainImage":null,"storyPackage":null,"contentPackage":null,"standout":{"editorsChoice":false,"exclusive":false,"scoop":false},"comments":{"enabled":true},"copyright":null,"webUrl":null,"publishReference":"tid_id_1","lastModified":"2017-03-30T13:09:06.480Z","canBeSyndicated":"verify","firstPublishedDate":"2017-03-30T13:08:53.000Z","accessLevel":"subscribed","canBeDistributed":"yes"},"contentUri":"http://wordpress-article-mapper/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	combiner := DummyDataCombiner{data: model.CombinedModel{UUID: "0cef259d-030d-497d-b4ef-e8fa0ee6db6b"}}
	producer := DummyMsgProducer{t: t, expError: errors.New("some producer error")}

	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Error sending transformed message to queue: %v", m.Headers["X-Request-Id"], producer.expError.Error()))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessContentMsg_Successfully_Forwarded(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"simple title"},"contentUri":"http://wordpress-article-mapper/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	combiner := DummyDataCombiner{data: model.CombinedModel{UUID: "0cef259d-030d-497d-b4ef-e8fa0ee6db6b", Content: model.ContentModel{UUID: "0cef259d-030d-497d-b4ef-e8fa0ee6db6b", Title: "simple title"}}}

	expMsg := producer.Message{
		Headers: m.Headers,
		Body:    `{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","content":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"simple title","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":false,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":null}`,
	}

	producer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", m.Headers["X-Request-Id"], combiner.data.UUID))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessContentMsg_DeleteEvent_Successfully_Forwarded(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1"},
		Body:    `{"payload":null,"contentUri":"http://wordpress-article-mapper/content/0cef259d-030d-497d-b4ef-e8fa0ee6db6b","lastModified":"2017-03-30T13:09:06.48Z"}`,
	}

	allowedUris := []string{"methode-article-mapper", "wordpress-article-mapper", "next-video-mapper"}
	config := MsgProcessorConfig{SupportedContentURIs: allowedUris}
	combiner := DummyDataCombiner{data: model.CombinedModel{
		UUID: "0cef259d-030d-497d-b4ef-e8fa0ee6db6b",
		Content: model.ContentModel{UUID: "0cef259d-030d-497d-b4ef-e8fa0ee6db6b",
			MarkedDeleted: true}}}

	expMsg := producer.Message{
		Headers: m.Headers,
		Body:    `{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","content":{"uuid":"0cef259d-030d-497d-b4ef-e8fa0ee6db6b","title":"","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":true,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":null}`,
	}

	producer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processContentMsg(m)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", m.Headers["X-Request-Id"], combiner.data.UUID))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessMetadataMsg_UnSupportedOrigins(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1", "Origin-System-Id": "origin"},
		Body:    `some body`,
	}

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}
	p := &MsgProcessor{config: config}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processMetadataMsg(m)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Skipped unsupported annotations with Origin-System-Id: %v. ", m.Headers["X-Request-Id"], m.Headers["Origin-System-Id"]))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessMetadataMsg_SupportedOrigin_Unmarshall_Error(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1", "Origin-System-Id": "http://cmdb.ft.com/systems/binding-service"},
		Body:    `some body`,
	}

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}
	p := &MsgProcessor{config: config}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processMetadataMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("Could not unmarshall message with TID=%v, error=", m.Headers["X-Request-Id"]))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessMetadataMsg_Combiner_Errors(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1", "Origin-System-Id": "http://cmdb.ft.com/systems/binding-service"},
		Body:    `{"uuid":"some_uuid","suggestions":[{"ID":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","PrefLabel":"Barclays","Types":["http://base-url/core/Thing","http://base-url/concept/Concept"],"Predicate":"http://base-url/about","ApiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","TmeIDs":["tme_id1"],"UUIDs":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"PlatformVersion":"v1"}]}`,
	}

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}
	combiner := DummyDataCombiner{err: errors.New("some error")}
	p := &MsgProcessor{config: config, DataCombiner: combiner}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processMetadataMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Error obtaining the combined message. Content couldn't get read. Message will be skipped.", m.Headers["X-Request-Id"]))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessMetadataMsg_Forwarder_Errors(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1", "Origin-System-Id": "http://cmdb.ft.com/systems/binding-service"},
		Body:    `{"uuid":"some_uuid","suggestions":[{"ID":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","PrefLabel":"Barclays","Types":["http://base-url/core/Thing","http://base-url/concept/Concept"],"Predicate":"http://base-url/about","ApiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","TmeIDs":["tme_id1"],"UUIDs":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"PlatformVersion":"v1"}]}`,
	}

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}
	combiner := DummyDataCombiner{data: model.CombinedModel{UUID: "some_uuid"}}
	producer := DummyMsgProducer{t: t, expError: errors.New("some producer error")}

	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processMetadataMsg(m)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Error sending transformed message to queue:", m.Headers["X-Request-Id"]))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestProcessMetadataMsg_Successfully_Forwarded(t *testing.T) {
	m := consumer.Message{
		Headers: map[string]string{"X-Request-Id": "some-tid1", "Origin-System-Id": "http://cmdb.ft.com/systems/binding-service"},
		Body:    `{"uuid":"some_uuid","suggestions":[{"ID":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","PrefLabel":"Barclays","Types":["http://base-url/core/Thing","http://base-url/concept/Concept"],"Predicate":"http://base-url/about","ApiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","TmeIDs":["tme_id1"],"UUIDs":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"PlatformVersion":"v1"}]}`,
	}

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}

	combiner := DummyDataCombiner{
		data: model.CombinedModel{
			UUID:    "some_uuid",
			Content: model.ContentModel{UUID: "some_uuid", Title: "simple title"},
			V1Metadata: []model.Annotation{
				{
					Thing: model.Thing{
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
						LeiCode:   "leicode_id_1",
						FactsetID: "factset-id1",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
				},
			},
		}}
	expMsg := producer.Message{
		Headers: m.Headers,
		Body:    `{"uuid":"some_uuid","content":{"uuid":"some_uuid","title":"simple title","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":false,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","leiCode":"leicode_id_1","factsetID":"factset-id1","tmeIDs":["tme_id1"],"uuids":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"platformVersion":"v1"}}]}`,
	}

	producer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	p.processMetadataMsg(m)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", m.Headers["X-Request-Id"], combiner.data.UUID))
	assert.Equal(t, 1, len(hook.Entries))
}

func TestForceMessage(t *testing.T) {

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}

	combiner := DummyDataCombiner{
		data: model.CombinedModel{
			UUID:    "some_uuid",
			Content: model.ContentModel{UUID: "some_uuid", Title: "simple title"},
			V1Metadata: []model.Annotation{
				{
					Thing: model.Thing{
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
						LeiCode:   "leicode_id_1",
						FactsetID: "factset-id1",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
				},
			},
		}}
	expMsg := producer.Message{
		Headers: map[string]string{"Message-Type": "cms-combined-content-published", "X-Request-Id": "[ignore]", "Origin-System-Id": "force-publish"},
		Body:    `{"uuid":"some_uuid","content":{"uuid":"some_uuid","title":"simple title","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":false,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":[{"thing":{"id":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","prefLabel":"Barclays","types":["http://base-url/core/Thing","http://base-url/concept/Concept","http://base-url/organisation/Organisation","http://base-url/company/Company","http://base-url/company/PublicCompany"],"predicate":"http://base-url/about","apiUrl":"http://base-url/80bec524-8c75-4d0f-92fa-abce3962d995","leiCode":"leicode_id_1","factsetID":"factset-id1","tmeIDs":["tme_id1"],"uuids":["80bec524-8c75-4d0f-92fa-abce3962d995","factset-generated-uuid"],"platformVersion":"v1"}}]}`,
	}

	producer := DummyMsgProducer{t: t, expUUID: combiner.data.UUID, expMsg: expMsg}
	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(combiner.data.UUID, "v1")
	assert.NoError(t, err)

	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, fmt.Sprintf("%v - Mapped and sent for uuid: %v", expMsg.Headers["X-Request-Id"], combiner.data.UUID))
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessageCombinerError(t *testing.T) {

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}
	combiner := DummyDataCombiner{err: errors.New("some error")}
	p := &MsgProcessor{config: config, DataCombiner: combiner}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(combiner.data.UUID, "v1")
	assert.Equal(t, combiner.err, err)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "Error obtaining the combined message, it will be skipped. some error")
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForceMessageProducerError(t *testing.T) {

	allowedOrigins := []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"}
	config := MsgProcessorConfig{SupportedHeaders: allowedOrigins}

	combiner := DummyDataCombiner{
		data: model.CombinedModel{
			UUID:    "some_uuid",
			Content: model.ContentModel{UUID: "some_uuid", Title: "simple title"},
			V1Metadata: []model.Annotation{
				{
					Thing: model.Thing{
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
						LeiCode:   "leicode_id_1",
						FactsetID: "factset-id1",
						TmeIDs:    []string{"tme_id1"},
						UUIDs: []string{"80bec524-8c75-4d0f-92fa-abce3962d995",
							"factset-generated-uuid"},
						PlatformVersion: "v1",
					},
				},
			},
		}}

	producer := DummyMsgProducer{t: t, expError: errors.New("some error")}
	p := &MsgProcessor{config: config, DataCombiner: combiner, MsgProducer: producer}

	hook := testLogger.NewGlobal()
	assert.Nil(t, hook.LastEntry())
	assert.Equal(t, 0, len(hook.Entries))

	err := p.ForceMessagePublish(combiner.data.UUID, "v1")
	assert.Equal(t, producer.expError, err)

	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "Error sending transformed message to queue: some error")
	assert.Equal(t, 2, len(hook.Entries))
}

func TestForwardMsg(t *testing.T) {

	tests := []struct {
		headers map[string]string
		uuid    string
		body    string
		err     error
	}{
		{
			headers: map[string]string{
				"X-Request-Id": "some-tid1",
			},
			uuid: "uuid1",
			body: `{"uuid":"uuid1","content":{"uuid":"","title":"","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":false,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":null}`,
			err:  nil,
		},
		{
			headers: map[string]string{
				"X-Request-Id": "some-tid2",
			},
			uuid: "uuid-returning-error",
			body: `{"uuid":"uuid-returning-error","content":{"uuid":"","title":"","body":"","identifiers":null,"publishedDate":"","lastModified":"","firstPublishedDate":"","mediaType":"","marked_deleted":false,"byline":"","standfirst":"","description":"","mainImage":"","publishReference":""},"v1-metadata":null}`,
			err:  fmt.Errorf("Some error"),
		},
	}

	for _, testCase := range tests {

		var model model.CombinedModel
		err := json.Unmarshal([]byte(testCase.body), &model)
		assert.Nil(t, err)

		q := MsgProcessor{
			MsgProducer: DummyMsgProducer{
				t:        t,
				expUUID:  testCase.uuid,
				expError: testCase.err,
				expMsg: producer.Message{
					Headers: testCase.headers,
					Body:    testCase.body,
				},
			},
		}

		err = q.forwardMsg(testCase.headers, &model)
		assert.Equal(t, testCase.err, err)
	}
}

func TestExtractTID(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		headers     map[string]string
		expectedTID string
	}{
		{
			headers: map[string]string{
				"X-Request-Id": "some-tid1",
			},
			expectedTID: "some-tid1",
		},
		{
			headers: map[string]string{
				"X-Request-Id":      "some-tid2",
				"Some-Other-Header": "some-value",
			},
			expectedTID: "some-tid2",
		},
	}

	for _, testCase := range tests {
		actualTID := extractTID(testCase.headers)
		assert.Equal(testCase.expectedTID, actualTID)
	}
}

func TestExtractTIDForEmptyHeader(t *testing.T) {
	assert := assert.New(t)
	headers := map[string]string{
		"Some-Other-Header": "some-value",
	}

	actualTID := extractTID(headers)
	assert.NotEmpty(actualTID)
	assert.Contains(actualTID, "tid_")
	assert.Contains(actualTID, "_post_publication_combiner")
}

func TestSupports(t *testing.T) {

	tests := []struct {
		element   string
		array     []string
		expResult bool
	}{
		{
			element:   "some value",
			array:     []string{"some value", "some other value"},
			expResult: true,
		},
		{
			element:   "some value",
			array:     []string{"some other value"},
			expResult: false,
		},
		{
			element:   "some value",
			array:     []string{},
			expResult: false,
		},
		{
			element:   "",
			array:     []string{"some value", "some other value"},
			expResult: false,
		},
		{
			element:   "",
			array:     []string{},
			expResult: false,
		},
	}

	for _, testCase := range tests {
		result := includes(testCase.array, testCase.element)
		assert.Equal(t, testCase.expResult, result, fmt.Sprintf("Element %v was not found in %v", testCase.array, testCase.element))
	}
}

func TestGetPlatformVersion(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		inputStr    string
		expectedStr string
	}{
		{"{some-url}/binding-service", "v1"},
		{"{some-url}/methode-web-pub", "v1"},
		{"methode-article-mapper", "v1"},
		{"wordpress-article-mapper", "v1"},
		{"next-video-mapper", "next-video"},
	}

	for _, testCase := range tests {
		pv := getPlatformVersion(testCase.inputStr)
		assert.Equal(testCase.expectedStr, pv)
	}
}

type DummyMsgProducer struct {
	t        *testing.T
	expUUID  string
	expMsg   producer.Message
	expError error
}

func (p DummyMsgProducer) SendMessage(uuid string, m producer.Message) error {

	if p.expError != nil {
		return p.expError
	}
	assert.Equal(p.t, m.Headers["Message-Type"], CombinerMessageType)
	assert.Equal(p.t, p.expUUID, uuid)
	if p.expMsg.Headers["X-Request-Id"] == "[ignore]" {
		p.expMsg.Headers["X-Request-Id"] = m.Headers["X-Request-Id"]
	}
	assert.Equal(p.t, p.expMsg, m)

	return nil
}

func (p DummyMsgProducer) ConnectivityCheck() (string, error) {
	return "", nil
}

type DummyDataCombiner struct {
	data model.CombinedModel
	err  error
}

func (c DummyDataCombiner) GetCombinedModelForContent(content model.ContentModel, platformVersion string) (model.CombinedModel, error) {
	return c.data, c.err
}

func (c DummyDataCombiner) GetCombinedModelForAnnotations(metadata model.Annotations, platformVersion string) (model.CombinedModel, error) {
	return c.data, c.err
}

func (c DummyDataCombiner) GetCombinedModel(uuid string, platformVersion string) (model.CombinedModel, error) {
	return c.data, c.err
}
