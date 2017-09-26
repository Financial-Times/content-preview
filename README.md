[![CircleCI](https://circleci.com/gh/Financial-Times/content-preview.svg?style=shield)](https://circleci.com/gh/Financial-Times/content-preview) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/content-preview/badge.svg)](https://coveralls.io/github/Financial-Times/content-preview)

# Content Preview Service (content-preview)

__Content Preview service serves published and unpublished content from Methode CMS (and in the future from Wordpress)
so that journalists can preview unpublished articles in the context that closely matches the live publication on Next FT platform.
It is one of internal services that comprise UPP read stack and as such is not meant to be accessible publicly
but to collaborate with other services such as Content-Public-Read-Preview service,
which in turn gives access to authenticated public API to read preview content.
Content Preview service relies on Methode API service and Methode Article Transformer service;
both underlying services are crucial to serve preview content__

## Installation

For the first time:

`go get github.com/Financial-Times/content-preview`

or update:

`go get -u github.com/Financial-Times/content-preview`

## Running


Locally with default configuration:

```
go install
$GOPATH/bin/content-preview
```

Locally with properties set:

```
go install
$GOPATH/bin/content-preview \
--app-port "8084" \
--source-app-name "methode-api" \
--source-app-auth "*<key>*" \
--source-app-uri "https://methodeapi.glb.ft.com/eom-file/" \
--source-app-health-uri "http://methode-api-uk-p.svc.ft.com/build-info" \
--source-app-panic-guide "https://sites.google.com/a/ft.com/dynamic-publishing-team/home/methode-api" \
--transform-app-name "methode-article-transformer" \
--transform-app-uri "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/map/" \
--transform-app-health-uri "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/build-info" \
--transform-app-panic-guide "https://sites.google.com/a/ft.com/dynamic-publishing-team/methode-article-transformer-panic-guide" \
--graphite-tcp-address "graphite.ft.com:2003" \
--graphite-prefix "coco.services.$ENV.content-preview.%i"
```

With Docker:

`docker build -t coco/content-preview .`

`docker run -ti coco/content-preview`

`docker run -ti
--env "APP_PORT=8080" \
--env "SOURCE_APP_NAME=methode-api" \
--env "SOURCE_APP_AUTH=*<key>*" \
--env "SOURCE_APP_URI=https://methodeapi.glb.ft.com/eom-file/" \
--env "SOURCE_APP_HEALTH_URI=https://methodeapi.glb.ft.com/build-info" \
--env "SOURCE_APP_PANIC_GUIDE=https://sites.google.com/a/ft.com/dynamic-publishing-team/home/methode-api" \
--env "TRANSFORM_APP_NAME=methode-article-transformer" \
--env "TRANSFORM_APP_URI=http://$HOSTNAME:8080/map/" \
--env "TRANSFORM_APP_HEALTH_URI=http://$HOSTNAME:8080/__health" \
--env "TRANSFORM_APP_PANIC_GUIDE=https://sites.google.com/a/ft.com/dynamic-publishing-team/methode-article-transformer-panic-guide" \
--env "GRAPHITE_TCP_ADDRESS=graphite.ft.com:2003" \
--env "GRAPHITE_PREFIX=coco.services.$ENV.content-preview.%i" \
coco/content-preview
`

When deployed locally arguments are optional.

##Authorization
SOURCE_APP_URI & SOURCE_APP_AUTH
When you set URI of the production instance of Methode API you also need to specify an authorisation key.
Authorisation key can be obtained from a secure note soted in LastPass.
You will need to ask team members to grant you access to the shared secure note containing production methode API key

## Endpoints
/content-preview/{uuid}    
Example
`curl -v http://localhost:8084/content-preview/9358ba1e-c07f-11e5-846f-79b0e3d20eaf`

### GET
The read should return transformed content of an article

404 if article with given uuid doesn't exist or cannot be transformed successfully

503 when one of the collaborating services is inaccessible


### Admin endpoints
Healthchecks: [http://localhost:8084/__health](http://localhost:8084/__health)

Ping: [http://localhost:8084/__ping](http://localhost:8084/__ping)

Build-info: [http://localhost:8084/__build-info](http://localhost:8084/__ping)  -  [Documentation on how to generate build-info] (https://github.com/Financial-Times/service-status-go)

Metrics:  [http://localhost:8084/__metrics](http://localhost:8084/__metrics)
