package main

import (
	"os"
	"github.com/jawher/mow.cli"
	log "github.com/Sirupsen/logrus"
	"regexp"
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

	sourceEnv := app.String(cli.StringOpt{
		Name:   "sourceEnv",
		Value:  "",
		Desc:   "Source environment (e.g. pub-xp)",
	})
	targetEnv := app.String(cli.StringOpt{
		Name:   "targetEnv",
		Value:  "",
		Desc:   "Target environment (e.g. xp)",
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
	sourceAuth := app.String(cli.StringOpt{
		Name:   "sourceAuth",
		Value:  "",
		Desc:   "Source credentials formatted as Basic auth header. (e.g. Basic abcdefg=)",
	})
	//targetAuth := app.String(cli.StringOpt{
	//	Name:   "targetAuth",
	//	Value:  "",
	//	Desc:   "targetCredentials formatted as Basic auth header. (e.g. Basic abcdefg=)",
	//})
	republishScope := app.String(cli.StringOpt{
		Name:   "republishScope",
		Value:  "",
		Desc:   "Republish scope (content, metadata, both)",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] publish-failure-resolver-go is starting ")

	app.Action = func() {
		log.Infof("%v", *sourceEnv)
		log.Infof("%v", *targetEnv)
		log.Infof("%v", *contentUuidsList)
		log.Infof("%v", *transactionIdPrefix)
		log.Infof("%v", *republishScope)

		httpClient := setupHttpClient()
		nativeStoreClient := NewNativeStoreClient(httpClient, "https://pub-xp-up.ft.com/__nativerw/", *sourceAuth)


		uuids := RegSplit(*contentUuidsList, "\\s")
		for _, uuid := range uuids {
			log.Infof("uuid=%v", uuid)
			for collection, _ := range collectionToOriginSystemId {
				//collection collectionToOriginSystemId[cms]

				if *republishScope == scopeBoth ||
					(collection == collectionV1Metadata && *republishScope == scopeMetadata) ||
					(collection != collectionV1Metadata && *republishScope == scopeContent) {
					nativeContent, err := nativeStoreClient.GetNative(collection, uuid, "tid_test")
					if err != nil {
						log.Warnf("can't publish uuid=%v %v", err)
					}
					log.Infof("%v", string(nativeContent))
				}
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
