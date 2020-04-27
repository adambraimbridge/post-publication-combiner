# upp-post-publication-combiner

[![CircleCI](https://circleci.com/gh/Financial-Times/post-publication-combiner.svg?style=svg)](https://circleci.com/gh/Financial-Times/post-publication-combiner)
[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/post-publication-combiner)](https://goreportcard.com/report/github.com/Financial-Times/post-publication-combiner)
[![Coverage Status](https://coveralls.io/repos/github/Financial-Times/post-publication-combiner/badge.svg)](https://coveralls.io/github/Financial-Times/post-publication-combiner)

## Introduction
This service builds combined messages (content + annotations) based on events received from `PostConceptAnnotations` or `PostPublicationEvents`.
This is a combination point for synchronizing the content and metadata publish flows.
Note: one publish event can result in two messages in the CombinedPostPublicationEvents topics (one for the content publish, and one for the metadata publish).

For `PostPublicationEvents` message the service extracts the published content from the messages, and requests the metadata from `public-annotations-api`. It is possible for `public-annotations-api`  to return `404 Not Found` for the provided content uuid.
For `PostConceptAnnotations` message the service extracts only the content uuid from the message, and requests the content from `document-store-api` and metadata from `public-annotations-api`. It is possible for `document-store-api` to return `404 Not Found` fot the provided content uuid.
The service then constructs a `CombinedPostPublicationEvents` message with the received data. It is possible for either `content` or `metadata` fields in the constructed message to be empty, but not both.

#### CombinedPostPublicationEvents format

```json5
{
  "uuid": "some_uuid", // content uuid
  "contentUri": "",
  "lastModified": "",
  "markedDeleted": "",
  "content": {}, // data returned from document-store-api
  "metadata": [] // data returned from public-annotations-api
}
```

### Dependencies 

- kafka/kafka-proxy
- document-store-api (/content endpoint)
- public-annotations-api (/content/{uuid}/annotations endpoint)

## Installation

In order to build, execute the following steps:

        go get github.com/Financial-Times/post-publication-combiner
        cd $GOPATH/src/github.com/Financial-Times/post-publication-combiner
        go build .

## Running locally

1. Run the tests and install the binary:

        go test ./...
        go install

1. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/post-publication-combiner

Please check --help for more details.

Test:
    You can verify the service's behavior by checking the consumed, and the generated kafka messages.
    You can also use the force endpoint.

## Build and deployment

* Built by Docker Hub (from master or from github tags): [coco/post-publication-combiner](https://hub.docker.com/r/coco/post-publication-combiner/)
* CI provided by CircleCI: [post-publication-combiner](https://circleci.com/gh/Financial-Times/post-publication-combiner)

## Service/Utility endpoints

### Force endpoint

`POST` - `/{content_uuid}` - Creates and forwards a CombinedPostPublicationEvent to the queue for the provided UUID.

Refer to [api.yml](_ft/api.yml) for api related documentation.

## Healthchecks
Our standard admin endpoints are:
`/__gtg` - returns 503 if any if the checks executed at the /__health endpoint returns false

`/__health`

Checks if:
* kafka-proxy is reachable
* document-store-api is reachable
* public-annotations-api is reachable

`/__build-info` 

### Logging

* The application uses the FT [go-logger](https://github.com/Financial-Times/go-logger) library, based on [logrus](https://github.com/sirupsen/logrus).
* NOTE: There is no logging for `/__build-info` and `/__gtg` endpoints as they are called frequently.