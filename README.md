# publish-failure-resolver-go

[![Circle CI](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/publish-failure-resolver-go)](https://goreportcard.com/report/github.com/Financial-Times/publish-failure-resolver-go) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/publish-failure-resolver-go/badge.svg)](https://coveralls.io/github/Financial-Times/publish-failure-resolver-go)

## Introduction

Reimport and/or republish content and lists in UPP.

Able to recognize image-sets if they are present in delivery's mongo, `upp-store/content`.

Parallelism and rate limiting configurable.

Written in Go.

## Installation

```
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
mkdir $GOPATH/src/github.com/Financial-Times/publish-failure-resolver-go
cd $GOPATH/src/github.com/Financial-Times/
git clone https://github.com/Financial-Times/publish-failure-resolver-go.git
cd publish-failure-resolver-go && dep ensure -vendor-only
go build .

go test ./...
```

## Running locally

```
./publish-failure-resolver-go \
  --sourceEnvHost="upp-staging-publish-eu.ft.com" \
  --targetEnvHost="upp-staging-publish-eu.ft.com" \
  --deliveryEnvHost="upp-staging-delivery-eu.ft.com" \
  --sourceAuth="username:password" \
  --targetAuth="username:password" \
  --deliveryAuth="username:password" \
  --republishScope="both" \
  --transactionIdPrefix="test76" \
  --rateLimitMs=200 \
  --parallelism=4 \
  --uuidList="ab36d158-f6cd-11e7-b6fb-5914dec7ca98 2316e87a-f084-11e7-892b-b579d79a9dbc 781a1047-3401-3df1-abf9-97b4a9e557d4 74d2df3c-f207-11e7-213f-3be68cc3546d aaaaaaaa-3d10-11e5-bbd1-bbbbbbbbbbbb 74d2df3c-f207-11e7-bf59-ac7c56b7ff24"
```

The options _rateLimit_, _parallelism_ and _scope_ are optional, the remaining are mandatory.

## Notes

Rate limit applies only to notifier endpoints, so searching in native-store and in upp-store are not considered rate limited actions.

The respective [jenkins job can be found here](http://ftjen06609-lvpr-uk-p:8181). Search for _Republish Failed Content and Metadata k8s go_
