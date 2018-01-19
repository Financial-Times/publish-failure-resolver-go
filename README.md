# publish-failure-resolver-go

[![Circle CI](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/publish-failure-resolver-go)](https://goreportcard.com/report/github.com/Financial-Times/publish-failure-resolver-go) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/publish-failure-resolver-go/badge.svg)](https://coveralls.io/github/Financial-Times/publish-failure-resolver-go)

## Introduction

Reimport and/or republish content and lists in UPP.

Able to recognize image-sets if they are present in delivery's mongo, `upp-store/content`.

Parallelism and rate limiting configurable.

Written in Go.

## Installation

```
go get -u github.com/kardianos/govendor
go get -u github.com/Financial-Times/publish-failure-resolver-go
cd $GOPATH/src/github.com/Financial-Times/publish-failure-resolver-go
govendor sync
go build .
```

## Running locally

```
./publish-failure-resolver-go --sourceEnvHost="pub-xp-up.ft.com" --targetEnvHost="pub-xp-up.ft.com" --deliveryEnvHost="xp-up.ft.com" --uuidList="674697de-fbb5-11e7-9b32-d7d59aace167 df6e71a2-d039-11e7-9dbb-291a884dd8c6 51bcfa75-5341-36e8-a2e8-b9f9d35d435f a7ad0dea-fc63-11e7-059a-92b661d49f6c" --transactionIdPrefix="test" --republishScope="both" --sourceAuth "Basic efgh" --targetAuth "Basic vxyz" --deliveryAuth "Basic abcd"
```

The options are all mandatory and they are self-explanatory, listed above.