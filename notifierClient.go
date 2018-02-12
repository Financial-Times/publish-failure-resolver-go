package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/sirupsen/logrus"
)

type notifierClient interface {
	Notify(nativeContent []byte, notifierApp, originSystemID, uuid, tid string) error
}

type httpNotifier struct {
	httpClient          *http.Client
	notifierAddressBase string
	authHeader          string
}

func newHTTPNotifier(httpClient *http.Client, notifierAddress, authHeader string) (*httpNotifier, error) {
	return &httpNotifier{
		httpClient:          httpClient,
		notifierAddressBase: notifierAddress,
		authHeader:          authHeader,
	}, nil
}

func (c *httpNotifier) Notify(nativeContent []byte, notifierApp, originSystemID, uuid, tid string) error {
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
	req.Header.Set("Connection", "close")
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
	io.Copy(ioutil.Discard, resp.Body)
	niceClose(resp)
	return nil
}
