# publish-failure-resolver-go

[![Circle CI](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/publish-failure-resolver-go)](https://goreportcard.com/report/github.com/Financial-Times/publish-failure-resolver-go) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/publish-failure-resolver-go/badge.svg)](https://coveralls.io/github/Financial-Times/publish-failure-resolver-go)

## Introduction

Republish, reimport or move content in UPP.

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
./publish-failure-resolver-go --sourceEnvHost="pub-xp-up.ft.com" --targetEnvHost="xp-up.ft.com" --contentUuidList="c301a10a-d058-11e7-b781-794ce08b24dc df6e71a2-d039-11e7-9dbb-291a884dd8c6 51bcfa75-5341-36e8-a2e8-b9f9d35d435f 51bcfa75-5341-36e8-a2e8-b9f9d35d435a" --transactionIdPrefix="tid_test" --republishScope="both" --sourceAuth "Basic abcdefg" --targetAuth "Basic abcded"
```

1. Run the tests and install the binary:

        govendor sync
        govendor test -v -race
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/publish-failure-resolver-go [--help]

Options:

        --app-system-code="publish-failure-resolver-go"            System Code of the application ($APP_SYSTEM_CODE)
        --app-name="publish-failure-resolver-go"                   Application name ($APP_NAME)
        --port="8080"                                           Port to listen on ($APP_PORT)
        
3. Test:

    ./publish-failure-resolver-go --sourcePropertiesFileLocation="" --targetPropertiesFileLocation="" --contentUuidList="d33af908-c8d9-11e7-357e-ed056da2bd77 8cef4c94-c8cd-11e7-357e-ed056da2bd77" --transactionIdPrefix="t123" --sourceCredentialsFileLocation="" --targetCredentialsFileLocation=""

## Build and deployment
_How can I build and deploy it (lots of this will be links out as the steps will be common)_

* Built by Docker Hub on merge to master: [publish-failure-resolver-go](https://hub.docker.com/r/publish-failure-resolver-go/)
* CI provided by CircleCI: [publish-failure-resolver-go](https://circleci.com/gh/Financial-Times/publish-failure-resolver-go)

## Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
