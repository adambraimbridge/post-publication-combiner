package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/Sirupsen/logrus"
	"github.com/golang/go/src/pkg/io"
	"io/ioutil"
	"net/http"
)

const (
	GTGEndpoint = "/__gtg"
)

type healthcheckHandler struct {
	httpClient                  http.Client
	proxyAddress                string
	proxyRequestHeader          string
	metadataTopic               string
	contentTopic                string
	combinedTopic               string
	docStoreApiBaseURL          string
	publicAnnotationsApiBaseURL string
}

func NewCombinerHealthcheck(proxyAddress string, proxyHeader string, client http.Client, metadataTopic string, contentTopic string, combinedTopic string, docStoreApiURL string, publicAnnotationsAPIURL string) *healthcheckHandler {
	return &healthcheckHandler{
		httpClient:                  client,
		proxyAddress:                proxyAddress,
		proxyRequestHeader:          proxyHeader,
		metadataTopic:               metadataTopic,
		contentTopic:                contentTopic,
		combinedTopic:               combinedTopic,
		docStoreApiBaseURL:          docStoreApiURL,
		publicAnnotationsApiBaseURL: publicAnnotationsAPIURL,
	}
}

func checkPostMetadataPublicationFoundHealthcheck(h *healthcheckHandler) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Metadata messages don't get processed. Content might not get indexed for search.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.metadataTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Metadata messages are not received from the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfPostMetadataPublicationTopicIsPresent,
	}
}

func checkPostContentPublicationTopicIsFoundHealthcheck(h *healthcheckHandler) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Content messages don't get processed. Content might not get indexed for search.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.contentTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Content messages are not received from the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfPostContentPublicationTopicIsPresent,
	}
}

func checkCombinedPublicationTopicTopicIsFoundHealthcheck(h *healthcheckHandler) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be written in the queue. Indexing for search won't work.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.combinedTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Messages couldn't be forwarded to the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfCombinedPublicationTopicIsPresent,
	}
}

func checkDocumentStoreApiHealthcheck(h *healthcheckHandler) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to document-store-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Document-store-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfDocumentStoreIsReachable,
	}
}

func checkPublicAnnotationsApiHealthcheck(h *healthcheckHandler) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to public-annotations-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Public-annotations-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfPublicAnnotationsApiIsReachable,
	}
}

func (h *healthcheckHandler) goodToGo(writer http.ResponseWriter, req *http.Request) {
	if _, err := h.checkIfPostContentPublicationTopicIsPresent(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
	if _, err := h.checkIfPostMetadataPublicationTopicIsPresent(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
	if _, err := h.checkIfCombinedPublicationTopicIsPresent(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
	if _, err := h.checkIfDocumentStoreIsReachable(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
	if _, err := h.checkIfPublicAnnotationsApiIsReachable(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *healthcheckHandler) checkIfPostMetadataPublicationTopicIsPresent() (string, error) {
	// TODO check what should the first return string represent - adjust logging/returns accordingly
	return "", checkIfTopicIsPresent(h, h.metadataTopic)
}

func (h *healthcheckHandler) checkIfPostContentPublicationTopicIsPresent() (string, error) {
	return "", checkIfTopicIsPresent(h, h.contentTopic)
}

func (h *healthcheckHandler) checkIfCombinedPublicationTopicIsPresent() (string, error) {
	return "", checkIfTopicIsPresent(h, h.combinedTopic)
}

func (h *healthcheckHandler) checkIfDocumentStoreIsReachable() (string, error) {
	b, err := utils.ExecuteSimpleHTTPRequest(h.docStoreApiBaseURL+GTGEndpoint, h.httpClient)
	return string(b), err
}

func (h *healthcheckHandler) checkIfPublicAnnotationsApiIsReachable() (string, error) {
	b, err := utils.ExecuteSimpleHTTPRequest(h.publicAnnotationsApiBaseURL+GTGEndpoint, h.httpClient)
	return string(b), err
}

func checkIfTopicIsPresent(h *healthcheckHandler, searchedTopic string) error {
	body, err := getTopicsFromProxy(h)
	if err != nil {
		logrus.Errorf("Healthcheck: Error reading request body: %v", err.Error())
		return err
	}

	var topics []string

	err = json.Unmarshal(body, &topics)
	if err != nil {
		logrus.Errorf("Connection could be established to kafka-proxy, but a parsing error occured and topic could not be found. %v", err.Error())
		return err
	}

	for _, topic := range topics {
		if topic == searchedTopic {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Connection could be established to kafka-proxy, but topic %s was not found", searchedTopic))
}

// Connect to the proxy by requesting the topic endpoint. Return the results.
func getTopicsFromProxy(h *healthcheckHandler) ([]byte, error) {
	req, err := http.NewRequest("GET", h.proxyAddress+"/topics", nil)
	if err != nil {
		logrus.Errorf("Error creating new kafka-proxy healthcheck request: %v", err.Error())
		return nil, err
	}

	if h.proxyRequestHeader != "" {
		req.Header.Add("Host",h.proxyRequestHeader)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("Healthcheck: Error executing kafka-proxy GET request: %v", err.Error())
		return nil, err
	}

	defer func() {
		_, err := io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			logrus.Warnf("message=\"couldn't read response body\" %v", err)
		}
		err = resp.Body.Close()
		if err != nil {
			logrus.Warnf("message=\"couldn't close response body\" %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Connecting to kafka proxy was not successful. Status: %d", resp.StatusCode))
	}
	return ioutil.ReadAll(resp.Body)
}
