package main

import (
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/post-publication-combiner/utils"
	"github.com/Financial-Times/service-status-go/gtg"
	log "github.com/Sirupsen/logrus"
)

const (
	GTGEndpoint            = "/__gtg"
	ResponseOK             = "OK"
	KafkaRestProxyEndpoint = "/__kafka-rest-proxy/topics"
)

type HealthcheckHandler struct {
	httpClient                  utils.Client
	proxyAddress                string
	proxyRequestHeader          string
	metadataTopic               string
	contentTopic                string
	combinedTopic               string
	docStoreAPIBaseURL          string
	publicAnnotationsAPIBaseURL string
}

func NewCombinerHealthcheck(proxyAddress string, proxyHeader string, client utils.Client, metadataTopic string, contentTopic string, combinedTopic string, docStoreAPIURL string, publicAnnotationsAPIURL string) *HealthcheckHandler {
	return &HealthcheckHandler{
		httpClient:                  client,
		proxyAddress:                proxyAddress,
		proxyRequestHeader:          proxyHeader,
		metadataTopic:               metadataTopic,
		contentTopic:                contentTopic,
		combinedTopic:               combinedTopic,
		docStoreAPIBaseURL:          docStoreAPIURL,
		publicAnnotationsAPIBaseURL: publicAnnotationsAPIURL,
	}
}

func checkKafkaProxyConnectivity(h *HealthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "Can't process PostPublicationEvents and PostMetadataPublicationEvents messages, can't write CombinedPostPublicationEvents messages to queue. Indexing for search won't work.",
		Name:             "Check connectivity to the kafka-proxy",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "PostPublicationEvents and PostMetadataPublicationEvents messages are not received from the queue, CombinedPostPublicationEvents messages can't be forwarded to the queue. Check if kafka-proxy is reachable.",
		Checker:          h.checkIfKafkaProxyIsReachable,
	}
}

func checkDocumentStoreAPIHealthcheck(h *HealthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to document-store-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Document-store-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfDocumentStoreIsReachable,
	}
}

func checkPublicAnnotationsAPIHealthcheck(h *HealthcheckHandler) health.Check {
	return health.Check{
		BusinessImpact:   "CombinedPostPublication messages can't be constructed. Indexing for content search won't work.",
		Name:             "Check connectivity to public-annotations-api",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/post-publication-combiner",
		Severity:         1,
		TechnicalSummary: "Public-annotations-api is not reachable. Messages can't be successfully constructed, neither forwarded.",
		Checker:          h.checkIfPublicAnnotationsAPIIsReachable,
	}
}

func (h *HealthcheckHandler) gtgCheck() gtg.Status {
	if _, err := h.checkIfKafkaProxyIsReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfDocumentStoreIsReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	if _, err := h.checkIfPublicAnnotationsAPIIsReachable(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}

	return gtg.Status{GoodToGo: true}
}

func (h *HealthcheckHandler) checkIfKafkaProxyIsReachable() (string, error) {
	urlStr := h.proxyAddress + KafkaRestProxyEndpoint

	_, _, err := utils.ExecuteSimpleHTTPRequest(urlStr, h.httpClient)
	if err != nil {
		log.Errorf("Healthcheck: %v", err.Error())
		return "", err
	}
	return ResponseOK, nil
}

func (h *HealthcheckHandler) checkIfDocumentStoreIsReachable() (string, error) {
	_, _, err := utils.ExecuteSimpleHTTPRequest(h.docStoreAPIBaseURL+GTGEndpoint, h.httpClient)
	if err != nil {
		log.Errorf("Healthcheck: %v", err.Error())
		return "", err
	}
	return ResponseOK, nil
}

func (h *HealthcheckHandler) checkIfPublicAnnotationsAPIIsReachable() (string, error) {
	_, _, err := utils.ExecuteSimpleHTTPRequest(h.publicAnnotationsAPIBaseURL+GTGEndpoint, h.httpClient)
	if err != nil {
		log.Errorf("Healthcheck: %v", err.Error())
		return "", err
	}
	return ResponseOK, nil
}
