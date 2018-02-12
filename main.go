package main

import (
	"encoding/base64"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

const (
	scopeBoth            = "both"
	scopeContent         = "content"
	scopeMetadata        = "metadata"
	cmsNotifier          = "cms-notifier"
	cmsMetadataNotifier  = "cms-metadata-notifier"
)

type targetSystem struct {
	name           string
	originSystemID string
	notifierApp    string
	scope          string
}

var defaultCollections = map[string]targetSystem{
	"methode": {
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
	"wordpress": {
		name:           "wordpress",
		originSystemID: "wordpress",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
	"video": {
		name:           "video",
		originSystemID: "next-video-editor",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
	"v1-metadata": {
		name:           "v1-metadata",
		originSystemID: "methode-web-pub",
		notifierApp:    cmsMetadataNotifier,
		scope:          scopeMetadata,
	},
	"next-video-editor": {
		name:           "video-metadata",
		originSystemID: "next-video-editor",
		notifierApp:    cmsMetadataNotifier,
		scope:          scopeMetadata,
	},
}

func main() {
	app := cli.App("publish-failure-resolver-go", "Republish, reimport or move content in UPP.")

	sourceEnvHost := app.String(cli.StringOpt{
		Name:  "sourceEnvHost",
		Value: "",
		Desc:  "Source environment's full hostname (e.g. upp-k8s-publishing-test-eu.ft.com or pub-xp-up.ft.com)",
	})
	targetEnvHost := app.String(cli.StringOpt{
		Name:  "targetEnvHost",
		Value: "",
		Desc:  "Target environment's full hostname (e.g. upp-k8s-delivery-test-eu.ft.com or xp-up.ft.com)",
	})
	deliveryEnvHost := app.String(cli.StringOpt{
		Name:  "deliveryEnvHost",
		Value: "",
		Desc:  "Delivery environment's full hostname, used for accessing document-store-api (e.g. upp-k8s-delivery-test-eu.ft.com or xp-up.ft.com)",
	})
	uuidList := app.String(cli.StringOpt{
		Name:  "uuidList",
		Value: "",
		Desc:  "Uuid list that you want to repbulish.",
	})
	transactionIDPrefix := app.String(cli.StringOpt{
		Name:  "transactionIdPrefix",
		Value: "",
		Desc:  "Transaction ID prefix",
	})
	sourceAuth := app.String(cli.StringOpt{
		Name:  "sourceAuth",
		Value: "",
		Desc:  "Source env credentials in format username:password (e.g. ops-01-01-2077:ABCDabcd)",
	})
	targetAuth := app.String(cli.StringOpt{
		Name:  "targetAuth",
		Value: "",
		Desc:  "Target env credentials in format username:password (e.g. ops-01-01-2077:ABCDabcd)",
	})
	deliveryAuth := app.String(cli.StringOpt{
		Name:  "deliveryAuth",
		Value: "",
		Desc:  "Delivery env credentials in format username:password (e.g. ops-01-01-2077:ABCDabcd)",
	})
	republishScope := app.String(cli.StringOpt{
		Name:  "republishScope",
		Value: "both",
		Desc:  "Republish scope (content, metadata, both)",
	})
	rateLimitMs := app.Int(cli.IntOpt{
		Name:  "rateLimitMs",
		Value: 1000,
		Desc:  "Rate limit at which one thread should not republish faster. (e.g. 200ms)",
	})
	parallelism := app.Int(cli.IntOpt{
		Name:  "parallelism",
		Value: 1,
		Desc:  "Number of parallel threads to take uuids and republish independently. must >= 1 (e.g. 16)",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] publish-failure-resolver-go is starting ")

	app.Action = func() {
		start := time.Now()

		log.Infof("sourceEnvHost=%v", *sourceEnvHost)
		log.Infof("targetEnvHost=%v", *targetEnvHost)
		log.Infof("deliveryEnvHost=%v", *deliveryEnvHost)
		log.Infof("transactionIdPrefix=%v", *transactionIDPrefix)
		log.Infof("republishScope=%v", *republishScope)
		log.Infof("rateLimitMs=%v", *rateLimitMs)
		log.Infof("parallelism=%v", *parallelism)

		httpClient := setupHTTPClient()
		nativeStoreClient := newNativeStoreClient(httpClient, "https://"+*sourceEnvHost+"/__nativerw/", "Basic "+base64.StdEncoding.EncodeToString([]byte(*sourceAuth)))
		notifierClient, err := newHTTPNotifier(httpClient, "https://"+*targetEnvHost+"/__", "Basic "+base64.StdEncoding.EncodeToString([]byte(*targetAuth)))
		docStoreClient, err := newHTTPDocStore(httpClient, "https://"+*deliveryEnvHost+"/__document-store-api/content", "Basic "+base64.StdEncoding.EncodeToString([]byte(*deliveryAuth)))
		rateLimit := time.Duration(*rateLimitMs) * time.Millisecond
		uuidCollectionRepublisher := newNotifyingUCRepublisher(notifierClient, nativeStoreClient, rateLimit)
		uuidRepublisher := newNotifyingUUIDRepublisher(uuidCollectionRepublisher, docStoreClient, defaultCollections)
		parallelRepublisher := newNotifyingParallelRepublisher(uuidRepublisher, *parallelism)
		if err != nil {
			log.Fatalf("Couldn't create notifier client. %v", err)
		}

		uuids := regSplit(*uuidList, "\\s")
		log.Infof("uuidList=%v", uuids)
		parallelRepublisher.Republish(uuids, *republishScope, *transactionIDPrefix)

		log.Infof("Dealt with nUuids=%v in duration=%v", len(uuids), time.Duration(time.Now().UnixNano()-start.UnixNano())*time.Nanosecond)
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func regSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(strings.TrimSpace(text), -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func setupHTTPClient() *http.Client {
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
