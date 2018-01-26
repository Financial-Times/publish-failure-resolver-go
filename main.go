package main

import (
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
	collectionV1Metadata = "v1-metadata"
	cmsNotifier          = "cms-notifier"
)

type targetSystem struct {
	name           string
	originSystemID string
	notifierApp    string
	scope          string
}

var collections = map[string]targetSystem{
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
		notifierApp:    "cms-metadata-notifier",
		scope:          scopeMetadata,
	},
	// "next-video-editor": {
	// 	name:           "next-video-editor",
	// 	originSystemID: "video-metadata",
	// 	notifierApp:    "cms-metadata-notifier",
	// 	scope:          scopeMetadata,
	// },
}

func main() {
	app := cli.App("publish-failure-resolver-go", "Republish, reimport or move content in UPP.")

	sourceEnvHost := app.String(cli.StringOpt{
		Name:  "sourceEnvHost",
		Value: "",
		Desc:  "Source environment's full hostname (e.g. pub-xp-up.ft.com or upp-k8s-publishing-test-eu.ft.com)",
	})
	targetEnvHost := app.String(cli.StringOpt{
		Name:  "targetEnvHost",
		Value: "",
		Desc:  "Target environment's full hostname (e.g. xp-up.ft.com or upp-k8s-delivery-test-eu.ft.com)",
	})
	deliveryEnvHost := app.String(cli.StringOpt{
		Name:  "deliveryEnvHost",
		Value: "",
		Desc:  "Delivery environment's full hostname, used for accessing document-store-api (e.g. xp-up.ft.com or upp-k8s-delivery-test-eu.ft.com)",
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
		Desc:  "Source env credentials formatted as Basic auth header. (e.g. Basic abcdefg=)",
	})
	targetAuth := app.String(cli.StringOpt{
		Name:  "targetAuth",
		Value: "",
		Desc:  "Target env credentials formatted as Basic auth header. (e.g. Basic abcdefg=)",
	})
	deliveryAuth := app.String(cli.StringOpt{
		Name:  "deliveryAuth",
		Value: "",
		Desc:  "Delivery env credentials formatted as Basic auth header. (e.g. Basic abcdefg=)",
	})
	republishScope := app.String(cli.StringOpt{
		Name:  "republishScope",
		Value: "",
		Desc:  "Republish scope (content, metadata, both)",
	})
	rateLimitMs := app.Int(cli.IntOpt{
		Name:  "rateLimitMs",
		Value: 200,
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
		log.Infof("sourceEnvHost=%v", *sourceEnvHost)
		log.Infof("targetEnvHost=%v", *targetEnvHost)
		log.Infof("deliveryEnvHost=%v", *deliveryEnvHost)
		log.Infof("transactionIdPrefix=%v", *transactionIDPrefix)
		log.Infof("republishScope=%v", *republishScope)
		log.Infof("rateLimitMs=%v", *rateLimitMs)
		log.Infof("parallelism=%v", *parallelism)

		httpClient := setupHTTPClient()
		nativeStoreClient := newNativeStoreClient(httpClient, "https://"+*sourceEnvHost+"/__nativerw/", *sourceAuth)
		notifierClient, err := newHTTPNotifier(httpClient, "https://"+*targetEnvHost+"/__", *targetAuth)
		docStoreClient, err := newHTTPDocStore(httpClient, "https://"+*deliveryEnvHost+"/__document-store-api/content", *deliveryAuth)
		republisher := newNotifyingRepublisher(notifierClient, docStoreClient, nativeStoreClient)
		parallelRepublisher := newNotifyingParallelRepublisher(republisher, *parallelism, time.Duration(*rateLimitMs)*time.Millisecond)
		if err != nil {
			log.Fatalf("Couldn't create notifier client. %v", err)
		}

		uuids := regSplit(*uuidList, "\\s")
		log.Infof("uuidList=%v", uuids)
		parallelRepublisher.Republish(uuids, *republishScope, *transactionIDPrefix)
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
