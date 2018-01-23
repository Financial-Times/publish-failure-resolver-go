package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/sirupsen/logrus"
)

type nativeStoreClientInterface interface {
	GetNative(collection, uuid, tid string) (nativeContent []byte, found bool, err error)
}

type nativeStoreClient struct {
	httpClient    *http.Client
	nativeAddress string
	authHeader    string
}

func newNativeStoreClient(httpClient *http.Client, nativeAddress, authHeader string) *nativeStoreClient {
	return &nativeStoreClient{
		httpClient:    httpClient,
		nativeAddress: nativeAddress,
		authHeader:    authHeader,
	}
}

func (c *nativeStoreClient) GetNative(collection, uuid, tid string) (nativeContent []byte, found bool, err error) {
	nativeURL, err := url.Parse(c.nativeAddress + collection + "/" + uuid)
	if err != nil {
		return nil, false, fmt.Errorf("invalid address nativeUrl=%v", nativeURL)
	}
	req, err := http.NewRequest(http.MethodGet, nativeURL.String(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("couldn't create request to fetch native content uuid=%v %v", uuid, err)
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	req.Header.Add("Authorization", c.authHeader)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("unsucessful request for fetching native content uuid=%v %v", uuid, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status while fetching native content uuid=%v collection=%v status=%v", uuid, collection, resp.StatusCode)
	}
	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	defer niceClose(resp)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read response body for uuid=%v %v", uuid, err)
	}
	ioutil.ReadAll(resp.Body)
	return bodyAsBytes, true, nil
}

func niceClose(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		logrus.Warnf("Couldn't close response body %v", err)
	}
}
