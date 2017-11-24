package main

import (
	"net/url"
	"fmt"
	"github.com/Financial-Times/transactionid-utils-go"
	"net/http"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

type nativeStoreClientInterface interface {
	GetNative(collection, uuid, tid string) (string, error)
}

type nativeStoreClient struct {
	httpClient 	  *http.Client
	nativeAddress string
	authHeader    string
}

func NewNativeStoreClient(httpClient *http.Client, nativeAddress, authHeader string) *nativeStoreClient {
	return &nativeStoreClient{
		httpClient:    httpClient,
		nativeAddress: nativeAddress,
		authHeader:    authHeader,
	}
}

func (c *nativeStoreClient) GetNative(collection, uuid, tid string) (nativeContent []byte, found bool, err error) {
	nativeUrl, err := url.Parse(c.nativeAddress + collection + "/" + uuid)
	if err != nil {
		return nil, false, fmt.Errorf("invalid address nativeUrl=%v", nativeUrl)
	}
	req, err := http.NewRequest(http.MethodGet, nativeUrl.String(), nil)
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
		return nil, false, fmt.Errorf("unexpected status while fetching native content uuid=%v collectoin=%v status=%v", uuid, collection, resp.StatusCode)
	}
	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	niceClose(resp)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read response body for uuid=%v %v", uuid, err)
	}
	return bodyAsBytes, true, nil
}

/*
	var content Content
	err = json.Unmarshal(bodyAsBytes, &content)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body for uuid=%v %v", uuid, err)
	}
 */

func niceClose(resp *http.Response) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logrus.Warnf("Couldn't close response body %v", err)
		}
	}()
}
