package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/stretchr/testify/assert"
)

const (
	DocStoreAPIPath          = "/doc-store-api"
	PublicAnnotationsAPIPath = "/public-annotations-api"
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
	assert.Empty(t, resp)
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
	assert.Empty(t, resp)
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

func TestGtgCheck_Good(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
	}
	h := HealthcheckHandler{
		httpClient:                  &dc,
		producerInstance:            &mockProducerInstance{isConnectionHealthy: true},
		docStoreAPIBaseURL:          "doc-store-base-url",
		publicAnnotationsAPIBaseURL: "pub-ann-base-url",
	}

	status := h.gtgCheck()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGtgCheck_Bad(t *testing.T) {
	testCases := []struct {
		description       string
		producerInstance  producer.MessageProducer
		docStoreAPIStatus int
		pubAnnAPIStatus   int
	}{
		{
			description:       "KafkaProxy GTG endpoint returns 503",
			producerInstance:  &mockProducerInstance{isConnectionHealthy: false},
			docStoreAPIStatus: 200,
			pubAnnAPIStatus:   200,
		},
		{
			description:       "DocumentStoreApi GTG endpoint returns 503",
			producerInstance:  &mockProducerInstance{isConnectionHealthy: true},
			docStoreAPIStatus: 503,
			pubAnnAPIStatus:   200,
		},
		{
			description:       "PublicAnnotationsApi GTG endpoint returns 503",
			producerInstance:  &mockProducerInstance{isConnectionHealthy: true},
			docStoreAPIStatus: 200,
			pubAnnAPIStatus:   503,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			server := getMockedServer(tc.docStoreAPIStatus, tc.pubAnnAPIStatus)
			defer server.Close()

			h := HealthcheckHandler{
				httpClient:                  http.DefaultClient,
				producerInstance:            tc.producerInstance,
				docStoreAPIBaseURL:          server.URL + DocStoreAPIPath,
				publicAnnotationsAPIBaseURL: server.URL + PublicAnnotationsAPIPath,
			}

			status := h.gtgCheck()
			assert.False(t, status.GoodToGo)
		})
	}
}

func getMockedServer(docStoreAPIStatus, pubAnnAPIStatus int) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc(DocStoreAPIPath+GTGEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(docStoreAPIStatus)
	})
	mux.HandleFunc(PublicAnnotationsAPIPath+GTGEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(pubAnnAPIStatus)
	})
	return server
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

type mockProducerInstance struct {
	isConnectionHealthy bool
}

func (p *mockProducerInstance) SendMessage(string, producer.Message) error {
	return nil
}

func (p *mockProducerInstance) ConnectivityCheck() (string, error) {
	if p.isConnectionHealthy {
		return "", nil
	}

	return "Error connecting to producer", errors.New("test")
}
