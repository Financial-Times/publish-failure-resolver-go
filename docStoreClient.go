package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
)

type docStoreClient interface {
	GetImageSetsModelUUID(setUUID, tid string) (found bool, modelUUID string, err error)
}

type httpDocStore struct {
	httpClient          *http.Client
	docStoreAddressBase string
	authHeader          string
}

func newHTTPDocStore(httpClient *http.Client, docStoreAddressBase, authHeader string) (*httpDocStore, error) {
	return &httpDocStore{
		httpClient:          httpClient,
		docStoreAddressBase: docStoreAddressBase,
		authHeader:          authHeader,
	}, nil
}

func (c *httpDocStore) GetImageSetsModelUUID(setUUID, tid string) (found bool, modelUUID string, err error) {
	docStoreURL, err := url.Parse(c.docStoreAddressBase + "/" + setUUID)
	if err != nil {
		return false, "", fmt.Errorf("coulnd't create URL for docStoreAddressBase=%v uuid=%v", c.docStoreAddressBase, setUUID)
	}
	req, err := http.NewRequest(http.MethodGet, docStoreURL.String(), nil)
	if err != nil {
		return false, "", fmt.Errorf("couldn't create request to get from document-store-api for uuid=%v %v", setUUID, err)
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	req.Header.Add("Authorization", c.authHeader)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("unsucessful request for getting uuid=%v %v", setUUID, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, "", nil
	}
	if resp.StatusCode != http.StatusOK {
		bodyAsBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("Couldn't read response body %v", err)
		}
		return false, "", fmt.Errorf("unexpected status while getting from document-store-api uuid=%v status=%v %v", setUUID, resp.StatusCode, string(bodyAsBytes))
	}
	defer niceClose(resp)

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Couldn't read response body %v", err)
	}
	return c.obtainImageModelUUID(bodyAsBytes, setUUID, tid)
}

func (c *httpDocStore) obtainImageModelUUID(bodyAsBytes []byte, setUUID, tid string) (found bool, modelUUID string, err error) {
	var content DocStoreContent
	jerr := json.Unmarshal(bodyAsBytes, &content)
	if err != nil {
		return false, "", fmt.Errorf("failed to unmarshal response body for uuid=%v tid=%v: %v", setUUID, tid, jerr.Error())
	}
	if content.Type != "ImageSet" {
		log.Warnf("found in document-store-api something that was not found in nativerw. Not as ImageSet but present. uuid=%v tid=%v", setUUID, tid)
		return false, "", nil
	}
	if len(content.Members) != 1 {
		return false, "", nil
	}
	return true, content.Members[0].UUID, nil
}
