package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"bytes"

	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
)

type notifierClientInterface interface {
	Notify(nativeContent []byte, uuid, tid string) error
}

type notifierClient struct {
	httpClient          *http.Client
	notifierAddressBase string
	authHeader          string
}

func newNotifierClient(httpClient *http.Client, notifierAddress, authHeader string) (*notifierClient, error) {
	return &notifierClient{
		httpClient:          httpClient,
		notifierAddressBase: notifierAddress,
		authHeader:          authHeader,
	}, nil
}

func (c *notifierClient) Notify(nativeContent []byte, notifierApp, originSystemID, uuid, tid string) error {
	notifierURL, err := url.Parse(c.notifierAddressBase + notifierApp + "/notify")
	if err != nil {
		return fmt.Errorf("coulnd't create URL for notifierAddressBase=%v notifierApp=%v", c.notifierAddressBase, notifierApp)
	}
	req, err := http.NewRequest(http.MethodPost, notifierURL.String(), bytes.NewReader(nativeContent))
	if err != nil {
		return fmt.Errorf("couldn't create request to notify for uuid=%v %v", uuid, err)
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	req.Header.Add("Authorization", c.authHeader)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Origin-System-Id", originSystemID)
	// log.Infof("Notify url=%v", notifierURL.String())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unsucessful request for notifying for uuid=%v %v", uuid, err)
	}
	if resp.StatusCode != http.StatusOK {
		bodyAsBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("Couldn't read response body %v", err)
		}
		return fmt.Errorf("unexpected status while notifying for uuid=%v status=%v %v", uuid, resp.StatusCode, string(bodyAsBytes))
	}
	niceClose(resp)
	return nil
}
