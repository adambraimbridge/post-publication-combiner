package main

import (
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/post-publication-combiner/processor"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
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

	// TODO log with transaction headers

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

		comb := processor.NewMessageCombiner(*docStoreApiBaseURL, *docStoreApiEndpoint, *publicAnnotationsApiBaseURL, *publicAnnotationsApiEndpoint, client)
		mf := processor.NewMessageForwarder(*kafkaProxyAddress, *kafkaProxyRoutingHeader, *combinedTopic, &client)

		cp := processor.NewContentQueueProcessor(*kafkaProxyAddress, *kafkaProxyRoutingHeader, *contentTopic, *kafkaConsumerGroup, *concurrentQueueProcessing, comb, mf)
		mp := processor.NewMetadataQueueProcessor(*kafkaProxyAddress, *kafkaProxyRoutingHeader, *metadataTopic, *kafkaConsumerGroup, *concurrentQueueProcessing, comb, mf)

		contentConsumer := consumer.NewConsumer(cp.Processor.QConf, cp.ProcessMsg, &client)
		metadataConsumer := consumer.NewConsumer(mp.Processor.QConf, mp.ProcessMsg, &client)

		go contentConsumer.Start()
		defer contentConsumer.Stop()

		go metadataConsumer.Start()
		defer metadataConsumer.Stop()

		// TODO too many parameters to pass -> refactor

		routeRequests(port, NewCombinerHealthcheck(*kafkaProxyAddress, *kafkaProxyRoutingHeader, client, *contentTopic, *metadataTopic, *combinedTopic, *docStoreApiBaseURL, *publicAnnotationsApiBaseURL))
	}

	logrus.SetLevel(logrus.InfoLevel)
	logrus.Infof("PostPublicationCombiner started with args %s", os.Args)
	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("App could not start, error=[%s]\n", err)
	}
}

func routeRequests(port *string, healthService *healthcheckHandler) {

	var monitoringRouter http.Handler = mux.NewRouter()
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(logrus.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	http.Handle("/", monitoringRouter)
	http.HandleFunc("/__health", v1a.Handler("Post-Publication-Combiner Healthcheck",
		"Checks for service dependencies: document-store, public-annotations-api, kafka proxy and the presence of related topics",
		checkPostMetadataPublicationFoundHealthcheck(healthService),
		checkPostContentPublicationTopicIsFoundHealthcheck(healthService),
		checkCombinedPublicationTopicTopicIsFoundHealthcheck(healthService),
		checkDocumentStoreApiHealthcheck(healthService),
		checkPublicAnnotationsApiHealthcheck(healthService)))
	http.HandleFunc("/__gtg", healthService.goodToGo)

	// TODO check if other endpoints are needed - build-info? fix it!

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		logrus.Fatalf("Unable to start: %v", err)
	}
}
