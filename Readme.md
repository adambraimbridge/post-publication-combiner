# post-publication-combiner

## Introduction
This service builds combined messages (content + v1 annotations) based on events received from PostMetadataPublicationEvents or PostPublicationEvents.  
The combined message is then sent to the CombinedPostPublicationEvents kafka queue.

This is a combination point for synchronizing the content and the metadata publish flows.
Note: one publish event can result in two messages in the CombinedPostPublicationEvents topics (one for the content publish, and one for the metadata publish).

This service depends on the following services:
- kafka/kafka-proxy
- document-store-api (/content endpoint)
- public-annotations-api (/content/{uuid}/annotations/{platformVersion} endpoint)

## Installation

In order to install, execute the following steps:

        ```go get -u github.com/Financial-Times/post-publication-combiner
        cd $GOPATH/src/github.com/Financial-Times/post-publication-combiner
        go get -t```

## Running locally

1. Run the tests and install the binary:

        go test ./...
        go install

1. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/post-publication-combiner [--help]
   
The available parameters are: 
* contentTopic
* metadataTopic
* combinedTopic
* kafkaProxyAddress
* kafkaConsumerGroup
* kafkaProxyHeader
* concurrent-queue-processing
* graphiteTCPAddress
* graphitePrefix
* logMetrics
* docStoreBaseURL - document-store-api base url (http://localhost:8080/__document-store-api)
* docStoreApiEndpoint - the endpoint for content retrieval (/content/{uuid})
* publicAnnotationsApiBaseURL - public-annotations-api base url (http://localhost:8080/__public-annotations-api)
* publicAnnotationsApiEndpoint - the endpoint for v1 metadata retrieval (/content/{uuid}/annotations/v1)

Please check --help for more details.

1. Test:
    For testing this service, please check the generated kafka messages.

## Build and deployment
How can I build and deploy it (lots of this will be links out as the steps will be common)

* Built by Docker Hub on merge to master: [coco/post-publication-combiner](https://hub.docker.com/r/coco/post-publication-combiner/)
* CI provided by CircleCI: [post-publication-combiner](https://circleci.com/gh/Financial-Times/post-publication-combiner)

## Service/Utility endpoints
This service does not provide any public endpoint (neither force), because the events should reflect the data received from any of the PostPublication events.
It makes no point to post something directly to the latter (combined) queue.

## Healthchecks
Our standard admin endpoints are:
`/\__gtg` - returns 503 if any if the checks executed at the /\__health endpoint returns false

`/\__health`

Checks if:
* kafka-proxy is reachable and PostMetadataPublicationEvents topic is present
* kafka-proxy is reachable and PostPublicationEvents topic is present
* kafka-proxy is reachable and CombinedPostPublicationEvents topic is present
* document-store-api is reachable
* public-annotations-api is reachable

`/build-info` (because our DW apps only have that, but with a planned migration to /\__build-info for new services, as we've done with our go apps)

### Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
* NOTE: `/build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.