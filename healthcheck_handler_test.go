package main

import (
	"errors"
	"fmt"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestCheckIfDocumentStoreIsReachable_Errors(t *testing.T) {

	expError := errors.New("some error")
	dc := dummyClient{
		err: expError,
	}
	h := healthcheckHandler{
		docStoreApiBaseURL: "doc-store-base-url",
		httpClient:         &dc,
	}

	resp, err := h.checkIfDocumentStoreIsReachable()
	assert.Contains(t, err.Error(), expError.Error(), fmt.Sprintf("Expected error %v not equal with recieved one %v", expError, err))
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfDocumentStoreIsReachable_Succeeds(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
		body:       "all good",
	}
	h := healthcheckHandler{
		docStoreApiBaseURL: "doc-store-base-url",
		httpClient:         &dc,
	}

	resp, err := h.checkIfDocumentStoreIsReachable()
	assert.Nil(t, err)
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfPublicAnnotationsApiIsReachable_Errors(t *testing.T) {

	expError := errors.New("some error")
	dc := dummyClient{
		err: expError,
	}
	h := healthcheckHandler{
		publicAnnotationsApiBaseURL: "pub-ann-base-url",
		httpClient:                  &dc,
	}

	resp, err := h.checkIfPublicAnnotationsApiIsReachable()
	assert.Contains(t, err.Error(), expError.Error(), fmt.Sprintf("Expected error %v not equal with recieved one %v", expError, err))
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfPublicAnnotationsApiIsReachable_Succeeds(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
		body:       "all good",
	}
	h := healthcheckHandler{
		publicAnnotationsApiBaseURL: "pub-ann-base-url",
		httpClient:                  &dc,
	}

	resp, err := h.checkIfPublicAnnotationsApiIsReachable()
	assert.Nil(t, err)
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfTopicIsPresent_HTTP_Call_Errors(t *testing.T) {

	tests := []struct {
		client   utils.Client
		topic    string
		expError error
	}{
		{
			dummyClient{
				err: errors.New("some error"),
			},
			"some-topic",
			errors.New("some error"),
		},
		{
			dummyClient{
				statusCode: http.StatusOK,
				body:       "cannot-be-unmarshalled",
			},
			"some-topic",
			errors.New("invalid character"),
		},
		{
			dummyClient{
				statusCode: http.StatusOK,
				body:       `["topic1","topic2"]`,
			},
			"some-topic",
			errors.New(fmt.Sprintf("Connection could be established to kafka-proxy, but topic %s was not found", "some-topic")),
		},
		{
			dummyClient{
				statusCode: http.StatusOK,
				body:       `["topic1","topic2"]`,
			},
			"topic1",
			nil,
		},
	}

	for _, testCase := range tests {
		h := healthcheckHandler{
			proxyAddress: "proxy-address",
			httpClient:   testCase.client,
		}
		err := checkIfTopicIsPresent(&h, testCase.topic)
		if testCase.expError == nil {
			assert.Nil(t, err)
		} else {
			assert.Contains(t, err.Error(), testCase.expError.Error(), fmt.Sprintf("Expected error %v not equal with recieved one %v", testCase.expError, err))
		}
	}
}

//func checkIfTopicIsPresent(h *healthcheckHandler, searchedTopic string) error {
//
//	urlStr := h.proxyAddress + "/__kafka-rest-proxy/topics"
//
//	body, _, err := utils.ExecuteSimpleHTTPRequest(urlStr, h.httpClient)
//	if err != nil {
//		logrus.Errorf("Healthcheck: %v", err.Error())
//		return err
//	}
//
//	var topics []string
//
//	err = json.Unmarshal(body, &topics)
//	if err != nil {
//		logrus.Errorf("Connection could be established to kafka-proxy, but a parsing error occurred and topic could not be found. %v", err.Error())
//		return err
//	}
//
//	for _, topic := range topics {
//		if topic == searchedTopic {
//			return nil
//		}
//	}
//
//	return errors.New(fmt.Sprintf("Connection could be established to kafka-proxy, but topic %s was not found", searchedTopic))
//}

type dummyClient struct {
	statusCode int
	body       string
	err        error
}

func (c dummyClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: c.statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(c.body)),
	}
	return resp, c.err
}
