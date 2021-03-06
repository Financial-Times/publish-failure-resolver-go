package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"

	"github.com/Financial-Times/publish-failure-resolver-go/pkg/image"
)

type DocStoreContent struct {
	Members []Member `json:"members"`
	Type    string   `json:"type"`
}

type Member struct {
	UUID string `json:"uuid"`
}

type httpDocStore struct {
	httpClient          *http.Client
	docStoreAddressBase string
	authHeader          string
}

func NewHTTPDocStore(httpClient *http.Client, docStoreAddressBase, authHeader string) (image.SetUUIDResolver, error) {
	return &httpDocStore{
		httpClient:          httpClient,
		docStoreAddressBase: docStoreAddressBase,
		authHeader:          authHeader,
	}, nil
}

func (c *httpDocStore) GetImageSetsModelUUID(setUUID, tid string) (bool, string, error) {
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
	req.Header.Set("Connection", "close")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("unsuccessful request for getting uuid=%v %v", setUUID, err)
	}
	defer niceClose(resp)
	if resp.StatusCode == http.StatusNotFound {
		io.Copy(ioutil.Discard, resp.Body)
		return false, "", nil
	}
	if resp.StatusCode != http.StatusOK {
		bodyAsBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("Couldn't read response body %v", err)
		}
		return false, "", fmt.Errorf("unexpected status while getting from document-store-api uuid=%v status=%v %v", setUUID, resp.StatusCode, string(bodyAsBytes))
	}

	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Couldn't read response body %v", err)
	}
	return c.obtainImageModelUUID(bodyAsBytes, setUUID, tid)
}

func (c *httpDocStore) obtainImageModelUUID(bodyAsBytes []byte, setUUID, tid string) (found bool, modelUUID string, err error) {
	var content DocStoreContent
	jerr := json.Unmarshal(bodyAsBytes, &content)
	if jerr != nil {
		return false, "", fmt.Errorf("failed to unmarshal response body for uuid=%v tid=%v: %v", setUUID, tid, jerr.Error())
	}
	if content.Type != "ImageSet" {
		log.Warnf("found in document-store-api something that was not found in nativerw. Not as ImageSet but present. uuid=%v tid=%v type=%v", setUUID, tid, content.Type)
		return false, "", nil
	}
	if len(content.Members) != 1 {
		log.Warnf("ImageSet is not of that type that has only one member. Those elements should be repbublished by using their own uuids, or enhancing this script to be able to.")
		return false, "", nil
	}
	return true, content.Members[0].UUID, nil
}
