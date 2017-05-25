package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	assert.Equal(t, "", resp)
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
	assert.Equal(t, "", resp)
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

func TestCheckIfKafkaProxyIsReachable_Succeeds(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
		body:       "all good",
	}
	h := HealthcheckHandler{
		proxyAddress: "kafka-proxy-address",
		httpClient:   &dc,
	}

	resp, err := h.checkIfKafkaProxyIsReachable()
	assert.Nil(t, err)
	assert.Equal(t, ResponseOK, resp)
}

func TestCheckIfKafkaProxyIsReachable_Errors(t *testing.T) {
	expError := errors.New("some error")
	dc := dummyClient{
		err: expError,
	}
	h := HealthcheckHandler{
		proxyAddress: "kafka-proxy-address",
		httpClient:   &dc,
	}

	resp, err := h.checkIfKafkaProxyIsReachable()
	assert.Contains(t, err.Error(), expError.Error(), fmt.Sprintf("Expected error %v not equal with recieved one %v", expError, err))
	assert.Equal(t, "", resp)
}

func TestGtgCheck_Good(t *testing.T) {
	dc := dummyClient{
		statusCode: http.StatusOK,
	}
	h := HealthcheckHandler{
		httpClient:                  &dc,
		docStoreAPIBaseURL:          "doc-store-base-url",
		proxyAddress:                "kafka-proxy-address",
		publicAnnotationsAPIBaseURL: "pub-ann-base-url",
	}

	status := h.gtgCheck()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGtgCheck_Bad(t *testing.T) {
	testCases := []struct {
		description       string
		kafkaProxyStatus  int
		docStoreAPIStatus int
		pubAnnAPIStatus   int
	}{
		{
			description:       "KafkaProxy GTG endpoint returns 503",
			kafkaProxyStatus:  503,
			docStoreAPIStatus: 200,
			pubAnnAPIStatus:   200,
		},
		{
			description:       "DocumentStoreApi GTG endpoint returns 503",
			kafkaProxyStatus:  200,
			docStoreAPIStatus: 503,
			pubAnnAPIStatus:   200,
		},
		{
			description:       "PublicAnnotationsApi GTG endpoint returns 503",
			kafkaProxyStatus:  200,
			docStoreAPIStatus: 200,
			pubAnnAPIStatus:   503,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			server := getMockedServer(tc.docStoreAPIStatus, tc.pubAnnAPIStatus, tc.kafkaProxyStatus)
			defer server.Close()

			h := HealthcheckHandler{
				httpClient:                  http.DefaultClient,
				docStoreAPIBaseURL:          server.URL + DocStoreAPIPath,
				proxyAddress:                server.URL,
				publicAnnotationsAPIBaseURL: server.URL + PublicAnnotationsAPIPath,
			}

			status := h.gtgCheck()
			assert.False(t, status.GoodToGo)
			assert.Contains(t, status.Message, "Status: 503")
		})
	}
}

func getMockedServer(docStoreAPIStatus, pubAnnAPIStatus, kafkaProxyStatus int) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc(DocStoreAPIPath+GTGEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(docStoreAPIStatus)
	})
	mux.HandleFunc(PublicAnnotationsAPIPath+GTGEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(pubAnnAPIStatus)
	})
	mux.HandleFunc(KafkaRestProxyEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(kafkaProxyStatus)
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
