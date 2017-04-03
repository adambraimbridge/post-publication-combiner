package main

import (
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/post-publication-combiner/processor"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {

	app := cli.App("post-publication-combiner", "Service listening to content and metadata PostPublication events, and forwards a combined message to the queue")

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	contentTopic := app.String(cli.StringOpt{
		Name:   "contentTopic",
		Value:  "PostPublicationEvents",
		EnvVar: "KAFKA_CONTENT_TOPIC_NAME",
	})
	metadataTopic := app.String(cli.StringOpt{
		Name:   "metadataTopic",
		Value:  "PostMetadataPublicationEvents",
		EnvVar: "KAFKA_METADATA_TOPIC_NAME",
	})
	combinedTopic := app.String(cli.StringOpt{
		Name:   "combinedTopic",
		Value:  "CombinedPostPublicationEvents",
		EnvVar: "KAFKA_COMBINED_TOPIC_NAME",
	})

	kafkaProxyAddress := app.String(cli.StringOpt{
		Name:   "kafkaProxyAddress",
		Value:  "http://localhost:8080",
		Desc:   "Address used by the queue consumer and producer to connect to the queue",
		EnvVar: "KAFKA_PROXY_ADDR",
	})
	kafkaConsumerGroup := app.String(cli.StringOpt{
		Name:   "kafkaConsumerGroup",
		Value:  "post-publication-combiner",
		Desc:   "Group used to read the messages from the queue",
		EnvVar: "KAFKA_PROXY_CONSUMER_GROUP",
	})
	kafkaProxyRoutingHeader := app.String(cli.StringOpt{
		Name:   "kafkaProxyHeader",
		Value:  "kafka",
		Desc:   "Kafka proxy header - used for vulcan routing.",
		EnvVar: "KAFKA_PROXY_HOST_HEADER",
	})
	concurrentQueueProcessing := app.Bool(cli.BoolOpt{
		Name:   "concurrent-queue-processing",
		Value:  false,
		Desc:   "Whether the consumers use concurrent processing for the messages",
		EnvVar: "KAFKA_PROXY_CONCURRENT_PROCESSING",
	})

	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.service-name.1 or content.test.people.rw.service-name.ftaps58938-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})

	docStoreApiBaseURL := app.String(cli.StringOpt{
		Name:   "docStoreApiBaseURL",
		Value:  "http://localhost:8080/__document-store-api",
		Desc:   "The address that the document store can be reached at. Important for content retrieval.",
		EnvVar: "DOCUMENT_STORE_BASE_URL",
	})
	docStoreApiEndpoint := app.String(cli.StringOpt{
		Name:   "docStoreApiEndpoint",
		Value:  "/content/{uuid}",
		Desc:   "The endpoint used for content retrieval.",
		EnvVar: "DOCUMENT_STORE_API_ENDPOINT",
	})
	publicAnnotationsApiBaseURL := app.String(cli.StringOpt{
		Name:   "publicAnnotationsApiBaseURL",
		Value:  "http://localhost:8080/__public-annotations-api",
		Desc:   "The address that the public-annotations-api can be reached at. Important for metadata retrieval.",
		EnvVar: "PUBLIC_ANNOTATIONS_API_BASE_URL",
	})
	publicAnnotationsApiEndpoint := app.String(cli.StringOpt{
		Name:   "publicAnnotationsApiEndpoint",
		Value:  "/content/{uuid}/annotations/v1",
		Desc:   "The endpoint used for metadata retrieval.",
		EnvVar: "PUBLIC_ANNOTATIONS_API_ENDPOINT",
	})
	whitelistedMetadataOriginSystemHeaders := app.Strings(cli.StringsOpt{
		Name:   "whitelistedMetadataOriginSystemHeaders",
		Value:  []string{"http://cmdb.ft.com/systems/binding-service", "http://cmdb.ft.com/systems/methode-web-pub"},
		Desc:   "Origin-System-Ids that are supported to be processed from the PostPublicationEvents queue.",
		EnvVar: "WHITELISTED_METADATA_ORIGIN_SYSTEM_HEADERS",
	})

	whitelistedContentUris := app.Strings(cli.StringsOpt{
		Name:   "whitelistedMetadataOriginSystemHeaders",
		Value:  []string{"methode-article-mapper", "wordpress-article-mapper", "brightcove-video-model-mapper"},
		Desc:   "Origin-System-Ids that are supported to be processed from the PostPublicationEvents queue.",
		EnvVar: "WHITELISTED_METADATA_ORIGIN_SYSTEM_HEADERS",
	})

	app.Action = func() {
		client := http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConnsPerHost:   20,
				TLSHandshakeTimeout:   3 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		contentConsumerConf := processor.NewQueueConsumerConfig(*kafkaProxyAddress, *kafkaProxyRoutingHeader, *contentTopic, *kafkaConsumerGroup, *concurrentQueueProcessing)
		metadataConsumerConf := processor.NewQueueConsumerConfig(*kafkaProxyAddress, *kafkaProxyRoutingHeader, *metadataTopic, *kafkaConsumerGroup, *concurrentQueueProcessing)
		producerConf := processor.NewProducerConfig(*kafkaProxyAddress, *combinedTopic, *kafkaProxyRoutingHeader)

		cp := processor.NewContentQueueProcessor(
			contentConsumerConf,
			producerConf,
			*publicAnnotationsApiBaseURL,
			*publicAnnotationsApiEndpoint,
			&client,
			*whitelistedContentUris,
		)
		cp.ProcessMessages()

		mp := processor.NewMetadataQueueProcessor(
			metadataConsumerConf,
			producerConf,
			*publicAnnotationsApiBaseURL,
			*publicAnnotationsApiEndpoint,
			*docStoreApiBaseURL,
			*docStoreApiEndpoint,
			&client,
			*whitelistedMetadataOriginSystemHeaders,
		)
		mp.ProcessMessages()

		routeRequests(port, NewCombinerHealthcheck(*kafkaProxyAddress, *kafkaProxyRoutingHeader, &client, *contentTopic, *metadataTopic, *combinedTopic, *docStoreApiBaseURL, *publicAnnotationsApiBaseURL))
	}

	logrus.SetLevel(logrus.InfoLevel)
	logrus.Infof("PostPublicationCombiner is starting with args %v", os.Args)

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("App could not start, error=[%v]\n", err)
	}
}

func routeRequests(port *string, healthService *healthcheckHandler) {

	r := mux.NewRouter()

	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(v1a.Handler("Post-Publication-Combiner Healthcheck",
		"Checks for service dependencies: document-store, public-annotations-api, kafka proxy and the presence of related topics",
		checkPostMetadataPublicationFoundHealthcheck(healthService),
		checkPostContentPublicationTopicIsFoundHealthcheck(healthService),
		checkCombinedPublicationTopicTopicIsFoundHealthcheck(healthService),
		checkDocumentStoreApiHealthcheck(healthService),
		checkPublicAnnotationsApiHealthcheck(healthService),
	))})
	r.Path("/__gtg").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(healthService.goodToGo)})

	if err := http.ListenAndServe(":"+*port, r); err != nil {
		logrus.Fatalf("Unable to start: %v", err)
	}
}
