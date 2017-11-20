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
	httpClient 	      *http.Client
	nativeAddress string
}

func NewNativeStoreClient(httpClient *http.Client, nativeAddress string) *nativeStoreClient {
	return &nativeStoreClient{
		httpClient:    httpClient,
		nativeAddress: nativeAddress,
	}
}

func (c *nativeStoreClient) GetNative(collection, uuid, tid string) ([]byte, error) {
	nativeUrl, err := url.Parse(c.nativeAddress + collection + "/" + uuid)
	if err != nil {
		return nil, fmt.Errorf("invalid address nativeUrl=%v", nativeUrl)
	}
	req, err := http.NewRequest(http.MethodGet, nativeUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request to fetch native content uuid=%v %v", uuid, err)
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsucessful request for fetching native content uuid=%v %v", uuid, err)
	}
	niceClose(resp)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("native content not found uuid=%v status=%v %v", uuid, resp.StatusCode, err)
	}
	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for uuid=%v %v", uuid, err)
	}
	return bodyAsBytes, nil
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
