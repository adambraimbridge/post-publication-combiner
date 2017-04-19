package main

import (
	"encoding/json"
	"fmt"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Sirupsen/logrus"
)

const (
	GTGEndpoint = "/__gtg"
	ResponseOK  = "OK"
)

type healthcheckHandler struct {
	httpClient                  utils.Client
	proxyAddress                string
	proxyRequestHeader          string
	metadataTopic               string
	contentTopic                string
	combinedTopic               string
	docStoreApiBaseURL          string
	publicAnnotationsApiBaseURL string
}

func NewCombinerHealthcheck(proxyAddress string, proxyHeader string, client utils.Client, metadataTopic string, contentTopic string, combinedTopic string, docStoreApiURL string, publicAnnotationsAPIURL string) *healthcheckHandler {
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

func checkPostMetadataPublicationFoundHealthcheck(h *healthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "Metadata messages don't get processed. Content might not get indexed for search.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.metadataTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Metadata messages are not received from the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfPostMetadataPublicationTopicIsPresent,
	}
}

func checkPostContentPublicationTopicIsFoundHealthcheck(h *healthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "Content messages don't get processed. Content might not get indexed for search.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.contentTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Content messages are not received from the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfPostContentPublicationTopicIsPresent,
	}
}

func checkCombinedPublicationTopicTopicIsFoundHealthcheck(h *healthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be written in the queue. Indexing for search won't work.",
		Name:             fmt.Sprintf("Check kafka-proxy connectivity and %s topic", h.combinedTopic),
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Messages couldn't be forwarded to the queue. Check if kafka-proxy is reachable and topic is present.",
		Checker:          h.checkIfCombinedPublicationTopicIsPresent,
	}
}

func checkDocumentStoreApiHealthcheck(h *healthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to document-store-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Document-store-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfDocumentStoreIsReachable,
	}
}

func checkPublicAnnotationsApiHealthcheck(h *healthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to public-annotations-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Public-annotations-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfPublicAnnotationsApiIsReachable,
	}
}

func (h *healthcheckHandler) gtgCheck() gtg.Status {
	if _, err := h.checkIfPostContentPublicationTopicIsPresent(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfPostMetadataPublicationTopicIsPresent(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfCombinedPublicationTopicIsPresent(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfDocumentStoreIsReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfPublicAnnotationsApiIsReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}

	return gtg.Status{GoodToGo: true}
}

func (h *healthcheckHandler) checkIfPostMetadataPublicationTopicIsPresent() (string, error) {
	return ResponseOK, checkIfTopicIsPresent(h, h.metadataTopic)
}

func (h *healthcheckHandler) checkIfPostContentPublicationTopicIsPresent() (string, error) {
	return ResponseOK, checkIfTopicIsPresent(h, h.contentTopic)
}

func (h *healthcheckHandler) checkIfCombinedPublicationTopicIsPresent() (string, error) {
	return ResponseOK, checkIfTopicIsPresent(h, h.combinedTopic)
}

func (h *healthcheckHandler) checkIfDocumentStoreIsReachable() (string, error) {
	_, _, err := utils.ExecuteSimpleHTTPRequest(h.docStoreApiBaseURL+GTGEndpoint, h.httpClient)
	if err != nil {
		logrus.Errorf("Healthcheck: %v", err.Error())
	}
	return ResponseOK, err
}

func (h *healthcheckHandler) checkIfPublicAnnotationsApiIsReachable() (string, error) {
	_, _, err := utils.ExecuteSimpleHTTPRequest(h.publicAnnotationsApiBaseURL+GTGEndpoint, h.httpClient)
	if err != nil {
		logrus.Errorf("Healthcheck: %v", err.Error())
	}
	return ResponseOK, err
}

func checkIfTopicIsPresent(h *healthcheckHandler, searchedTopic string) error {

	urlStr := h.proxyAddress + "/__kafka-rest-proxy/topics"

	body, _, err := utils.ExecuteSimpleHTTPRequest(urlStr, h.httpClient)
	if err != nil {
		logrus.Errorf("Healthcheck: %v", err.Error())
		return err
	}

	var topics []string

	err = json.Unmarshal(body, &topics)
	if err != nil {
		logrus.Errorf("Connection could be established to kafka-proxy, but a parsing error occurred and topic could not be found. %v", err.Error())
		return err
	}

	for _, topic := range topics {
		if topic == searchedTopic {
			return nil
		}
	}

	return fmt.Errorf("Connection could be established to kafka-proxy, but topic %s was not found", searchedTopic)
}
