# UPP - Post Publication Combiner

Post Publication Combiner is a service that combines content and annotations publish events in a single event.

## Code

post-publication-combiner

## Primary URL

https://upp-prod-delivery-glb.upp.ft.com/__post-publication-combiner/

## Service Tier

Platinum

## Lifecycle Stage

Production

## Delivered By

content

## Supported By

content

## Known About By

- elitsa.pavlova
- kalin.arsov
- miroslav.gatsanoga
- ivan.nikolov
- marina.chompalova
- hristo.georgiev
- elina.kaneva
- georgi.kazakov

## Host Platform

AWS

## Architecture

This service builds combined messages (content + annotations) based on events received from PostConceptAnnotations or PostPublicationEvents. The combined message is then sent to the CombinedPostPublicationEvents Kafka queue, later being consumed by the Elasticsearch indexer.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Dependencies

- upp-kafka
- document-store-api
- annotationsapi

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both Delivery clusters. The failover guide for the cluster is located here:
https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

If the new release does not change the way kafka messages are consumed and/or produce it's safe to deploy it without cluster failover.

## Key Management Process Type

Manual

## Key Management Details

To access the service clients need to provide basic auth credentials.
To rotate credentials you need to login to a particular cluster and update varnish-auth secrets.

## Monitoring

Service in UPP K8S delivery clusters:

- Delivery-Prod-EU health: https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=post-publication-combiner
- Delivery-Prod-US health: https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=post-publication-combiner

## First Line Troubleshooting

https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.
