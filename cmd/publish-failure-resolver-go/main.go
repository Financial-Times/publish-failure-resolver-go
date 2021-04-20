package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"

	"github.com/Financial-Times/publish-failure-resolver-go/pkg/http"
	"github.com/Financial-Times/publish-failure-resolver-go/pkg/http/api"
	"github.com/Financial-Times/publish-failure-resolver-go/pkg/image"
	"github.com/Financial-Times/publish-failure-resolver-go/pkg/republisher"
)

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
		Desc:  "Uuid list that you want to republish.",
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
	denylistPath := app.String(cli.StringOpt{
		Name:  "denylist",
		Value: "/denylist.txt",
		Desc:  "Path to UUID collection which are denied from updating.",
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
		log.Infof("denylistPath=%v", *denylistPath)

		httpClient := http.NewHTTPClient()
		nativeStoreClient := api.NewNativeStoreClient(httpClient, "https://"+*sourceEnvHost+"/__nativerw/", "Basic "+base64.StdEncoding.EncodeToString([]byte(*sourceAuth)))
		notifierClient, err := api.NewHTTPNotifier(httpClient, "https://"+*targetEnvHost+"/__", "Basic "+base64.StdEncoding.EncodeToString([]byte(*targetAuth)))
		var imageSetResolver image.SetUUIDResolver
		if *deliveryEnvHost == "" || *deliveryAuth == "" {
			imageSetResolver = image.NewUUIDImageSetResolver()
		} else {
			imageSetResolver, err = api.NewHTTPDocStore(httpClient, "https://"+*deliveryEnvHost+"/__document-store-api/content", "Basic "+base64.StdEncoding.EncodeToString([]byte(*deliveryAuth)))
		}
		rateLimit := time.Duration(*rateLimitMs) * time.Millisecond
		uuidCollectionRepublisher := republisher.NewNotifyingUCRepublisher(notifierClient, nativeStoreClient, rateLimit)
		uuidRepublisher := republisher.NewNotifyingUUIDRepublisher(uuidCollectionRepublisher, imageSetResolver, republisher.DefaultCollections)
		var r republisher.BulkRepublisher
		if *parallelism > 1 {
			r = republisher.NewNotifyingParallelRepublisher(uuidRepublisher, *parallelism)
		} else {
			r = republisher.NewNotifyingSequentialRepublisher(uuidRepublisher)
		}

		if err != nil {
			log.Fatalf("Couldn't create notifier client. %v", err)
		}

		uuids := regSplit(*uuidList)

		denylistedUUIDs, err := readDenylistedUUIDs(*denylistPath)
		if err != nil {
			log.Fatalf("Couldn't read deny-listed UUIDs. %v", err)
		}

		uuids = removeDenylistedUUIDs(uuids, denylistedUUIDs)

		log.Infof("uuidList=%v", uuids)
		_, errs := r.Republish(uuids, *republishScope, *transactionIDPrefix)

		log.Infof("Dealt with nUuids=%v in duration=%v", len(uuids), time.Duration(time.Now().UnixNano()-start.UnixNano())*time.Nanosecond)

		if len(errs) > 0 {
			os.Exit(1)
		}
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

// Returns a list of all deny-listed from republishing UUIDs.
func readDenylistedUUIDs(denylistPath string) ([]string, error) {
	file, err := os.Open(denylistPath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var denylistedUUIDs []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		denylistedUUIDs = append(denylistedUUIDs, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning error: %w", err)
	}

	return denylistedUUIDs, nil
}

// Filters out deny-listed UUIDs from scheduled for republishing ones.
func removeDenylistedUUIDs(republishUUIDs, denylistedUUIDs []string) []string {
	denylisted := make(map[string]struct{}, len(denylistedUUIDs))

	for _, uuid := range denylistedUUIDs {
		denylisted[uuid] = struct{}{}
	}

	filteredUUIDs := make([]string, 0, len(republishUUIDs))

	for _, uuid := range republishUUIDs {
		if _, ok := denylisted[uuid]; !ok {
			filteredUUIDs = append(filteredUUIDs, uuid)
		}
	}

	return filteredUUIDs
}

func regSplit(text string) []string {
	reg := regexp.MustCompile(`\s`)
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
