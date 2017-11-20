package main

import (
	"os"
	"github.com/jawher/mow.cli"
	log "github.com/Sirupsen/logrus"
	"regexp"
	"net/url"
	"fmt"
	"github.com/Financial-Times/transactionid-utils-go"
	"net"
	"time"
	"net/http"
)

const (
	scopeBoth = "both"
	scopeContent = "content"
	scopeMetadata = "metadata"
	collectionV1Metadata = "v1-metadata"
)

var collectionToOriginSystemId = map[string]string {
	"v1-metadata": "methode",
}

func main() {
	app := cli.App("publish-failure-resolver-go", "Republish, reimport or move content in UPP.")

	sourcePropertiesFileLocation := app.String(cli.StringOpt{
		Name:   "sourcePropertiesFileLocation",
		Value:  "",
		Desc:   "Source properties file",
	})
	targetPropertiesFileLocation := app.String(cli.StringOpt{
		Name:   "targetPropertiesFileLocation",
		Value:  "",
		Desc:   "Target properties file",
	})
	contentUuidsList := app.String(cli.StringOpt{
		Name:   "contentUuidList",
		Value:  "",
		Desc:   "Content uuid list",
	})
	transactionIdPrefix := app.String(cli.StringOpt{
		Name:   "transactionIdPrefix",
		Value:  "",
		Desc:   "Transaction ID prefix",
	})
	sourceCredentialsFileLocation := app.String(cli.StringOpt{
		Name:   "sourceCredentialsFileLocation",
		Value:  "",
		Desc:   "Source credentials file location",
	})
	targetCredentialsFileLocation := app.String(cli.StringOpt{
		Name:   "targetCredentialsFileLocation",
		Value:  "",
		Desc:   "Target credentials file location",
	})
	republishScope := app.String(cli.StringOpt{
		Name:   "republishScope",
		Value:  "",
		Desc:   "Republish scope (content, metadata, both)",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] publish-failure-resolver-go is starting ")

	app.Action = func() {
		log.Infof("%v", *sourcePropertiesFileLocation)
		log.Infof("%v", *targetPropertiesFileLocation)
		log.Infof("%v", *contentUuidsList)
		log.Infof("%v", *transactionIdPrefix)
		log.Infof("%v", *sourceCredentialsFileLocation)
		log.Infof("%v", *targetCredentialsFileLocation)

		httpClient := setupHttpClient()
		NewNativeStoreClient(httpClient, "https://pub-xp-up.ft.com/__nativerw/")

		uuids := RegSplit(*contentUuidsList, "\\s")
		for _, uuid := range uuids {
			log.Infof("uuid=%v", uuid)
			for cms, _ := range collectionToOriginSystemId {
				//collectionToOriginSystemId[cms]

				if *republishScope == scopeBoth ||
					(cms == collectionV1Metadata && *republishScope == scopeMetadata) ||
					(cms != collectionV1Metadata && *republishScope == scopeContent) {

				}
				getContentResult=`curl -qSfs -u ${sourceCredentials} -H "X-Request-Id : ${transactionHeader}" https://${sourceCluster}/__${sourceService}/${cms}/${uuid} 2>/dev/null`
			}
		}
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes) + 1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func setupHttpClient() *http.Client {
	return &http.Client{
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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
