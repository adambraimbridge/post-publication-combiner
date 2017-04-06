package processor

import (
	"encoding/json"
	"errors"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/post-publication-combiner/model"
	"github.com/golang/go/src/pkg/fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

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

type DummyConsumer struct {
}

func (DummyConsumer) Start() {
}

func (DummyConsumer) Stop() {
}

func (DummyConsumer) ConnectivityCheck() (string, error) {
	return "", nil
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
	assert.True(p.t, reflect.DeepEqual(p.expMsg, m))

	return nil
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
			err:  errors.New(fmt.Sprint("Some error")),
		},
	}

	for _, testCase := range tests {

		var model model.CombinedModel
		err := json.Unmarshal([]byte(testCase.body), &model)
		assert.Nil(t, err)

		q := QueueProcessor{
			MessageConsumer: DummyConsumer{},
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

func TestContains(t *testing.T) {

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
		result := contains(testCase.array, testCase.element)
		assert.Equal(t, testCase.expResult, result, fmt.Sprintf("Element %v was not found in %v", testCase.array, testCase.element))
	}
}
