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
	h := HealthcheckHandler{
		docStoreAPIBaseURL: "doc-store-base-url",
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
	h := HealthcheckHandler{
		docStoreAPIBaseURL: "doc-store-base-url",
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
	h := HealthcheckHandler{
		publicAnnotationsAPIBaseURL: "pub-ann-base-url",
		httpClient:                  &dc,
	}

	resp, err := h.checkIfPublicAnnotationsAPIIsReachable()
	assert.Contains(t, err.Error(), expError.Error(), fmt.Sprintf("Expected error %v not equal with recieved one %v", expError, err))
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfPublicAnnotationsApiIsReachable_Succeeds(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
		body:       "all good",
	}
	h := HealthcheckHandler{
		publicAnnotationsAPIBaseURL: "pub-ann-base-url",
		httpClient:                  &dc,
	}

	resp, err := h.checkIfPublicAnnotationsAPIIsReachable()
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
			fmt.Errorf("Connection could be established to kafka-proxy, but topic %s was not found", "some-topic"),
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
		h := HealthcheckHandler{
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
