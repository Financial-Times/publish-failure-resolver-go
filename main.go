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
	"methode": "methode-web-pub",
	"wordpress": "wordpress",
	"video": "next-video-editor",
	"v1-metadata": "methode-web-pub",
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
		log.Infof("sourceEnv=%v", *sourceEnv)
		log.Infof("targetEnv=%v", *targetEnv)
		log.Infof("contentUuidsList=%v", *contentUuidsList)
		log.Infof("transactionIdPrefix=%v", *transactionIdPrefix)
		log.Infof("republishScope=%v", *republishScope)

		httpClient := setupHttpClient()
		nativeStoreClient := NewNativeStoreClient(httpClient, "https://pub-xp-up.ft.com/__nativerw/", *sourceAuth)

		uuids := RegSplit(*contentUuidsList, "\\s")
		for _, uuid := range uuids {
			log.Infof("uuid=%v", uuid)
			isFoundInAnyCollection := false
			var nativeContent []byte
			for collection, _ := range collectionToOriginSystemId {
				if *republishScope == scopeBoth ||
					(collection == collectionV1Metadata && *republishScope == scopeMetadata) ||
					(collection != collectionV1Metadata && *republishScope == scopeContent) {
					var err error
					var isFound bool
					nativeContent, isFound, err = nativeStoreClient.GetNative(collection, uuid, "tid_test")
					if err != nil {
						log.Warnf("error while fetching native content: %v", err)
						continue
					}
					if !isFound {
						continue

					}
					isFoundInAnyCollection = true
					originSystemId := collectionToOriginSystemId[collection]
					log.Infof("found uuid=%v in collection=%v originSystemId=%v", uuid, collection, originSystemId)
					log.Infof("%v", string(nativeContent))
				}
			}
			if !isFoundInAnyCollection {
				log.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections", uuid)
				continue
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
