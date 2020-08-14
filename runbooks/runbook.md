# UPP - Content Preview

Content Preview serves published and unpublished content from Methode CMS so that journalists can preview unpublished articles.

## Code

content-preview

## Primary URL

https://upp-prod-delivery-glb.upp.ft.com/__content-preview/

## Service Tier

Bronze

## Lifecycle Stage

Production

## Delivered By

content

## Supported By

content

## Known About By

- dimitar.terziev
- elitsa.pavlova
- kalin.arsov
- hristo.georgiev
- tsvetan.dimitrov
- elina.kaneva
- robert.marinov
- georgi.ivanov

## Host Platform

AWS

## Architecture

The Content Preview service serves published and unpublished content from Methode CMS so that journalists can preview unpublished articles in the context that closely matches the live publication on the Next FT platform. 
It is one of internal services that comprise UPP read stack and is not meant to be accessible publicly but to collaborate with other services such as `content-public-read-preview`, which in turn gives access to an authenticated public API to read preview content. 
Content Preview relies on Methode API service and Methode Article Transformer service; both underlying services are crucial to serve preview content.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Dependencies

- up-mam
- up-maicm
- up-mapi

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both Delivery clusters. The failover guide for the cluster is located here: https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster.

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

The release is triggered by making a Github release which is then picked up by a Jenkins multibranch pipeline. The Jenkins pipeline should be manually started in order for it to deploy the helm package to the Kubernetes clusters.

## Key Management Process Type

NotApplicable

## Key Management Details

There is no key rotation procedure for this system.

## Monitoring

Service in UPP K8S delivery clusters:

- Delivery-Prod-EU health: https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=content-preview
- Delivery-Prod-US health: https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=content-preview

## First Line Troubleshooting

[First Line Troubleshooting guide](https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting)

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.
