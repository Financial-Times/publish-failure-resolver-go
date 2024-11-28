# publish-failure-resolver-go

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/Financial-Times/publish-failure-resolver-go/tree/master.svg?style=svg&circle-token=CCIPRJ_AafnQ3k8Cj7ofMuCTuWtZa_ddab9cb4d80553371a0fa13d17c02811554d349e)] (https://dl.circleci.com/status-badge/redirect/gh/Financial-Times/publish-failure-resolver-go/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/publish-failure-resolver-go)](https://goreportcard.com/report/github.com/Financial-Times/publish-failure-resolver-go) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/publish-failure-resolver-go/badge.svg?branch=feature/UPPSF-1102-duplicate-uuids)](https://coveralls.io/github/Financial-Times/publish-failure-resolver-go?branch=master)

## Introduction

Reimport and/or republish content and lists in UPP.

Parallelism and rate limiting configurable.

## Installation

```
git clone https://github.com/Financial-Times/publish-failure-resolver-go.git
cd $GOPATH/src/github.com/Financial-Times/publish-failure-resolver-go
go build -mod=readonly .
```

## Running locally

```
./publish-failure-resolver-go \
  --sourceEnvHost="upp-staging-publish-eu.ft.com" \
  --targetEnvHost="upp-staging-publish-eu.ft.com" \
  --sourceAuth="username:password" \
  --targetAuth="username:password" \
  --republishScope="both" \
  --transactionIdPrefix="test76" \
  --rateLimitMs=200 \
  --parallelism=4 \
  --uuidList="ab36d158-f6cd-11e7-b6fb-5914dec7ca98 2316e87a-f084-11e7-892b-b579d79a9dbc 781a1047-3401-3df1-abf9-97b4a9e557d4 74d2df3c-f207-11e7-213f-3be68cc3546d aaaaaaaa-3d10-11e5-bbd1-bbbbbbbbbbbb 74d2df3c-f207-11e7-bf59-ac7c56b7ff24"
```

The options _rateLimit_, _parallelism_ and _scope_ are optional, the remaining are mandatory.

## Running the tests                  

```shell
go test -mod=readonly -race ./...
```

## Build and deployment

- Built by Docker Hub on merge to master: [coco/publish-failure-resolver-go](https://hub.docker.com/r/coco/publish-failure-resolver-go/)
- CI provided by CircleCI: [publish-failure-resolver-go](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go)

## Local docker build

```sh
docker build -t publish-failure-resolver-go:local .
```

### Logging

* The application uses [logrus](https://github.com/sirupsen/logrus); the log file is initialised in [main.go](cmd/publish-failure-resolver-go/main.go).

## Notes

Rate limit applies only to notifier endpoints, so searching in native-store and in upp-store are not considered rate limited actions.

The respective [jenkins job can be found here](https://upp-jenkins-k8s-prod.upp.ft.com/job/publish-utils/job/OLD%20-%20For%20Developers%20-%20Republish%20Failed%20Content%20and%20Metadata/), or if searched for _Republish Failed Content and Metadata k8s go_
