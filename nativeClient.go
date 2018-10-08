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

type nativeMSG struct {
	body           []byte
	contentType    string
	originSystemID string
}

type nativeStoreClientInterface interface {
	GetNative(collection, uuid, tid string) (nativeContent *nativeMSG, found bool, err error)
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

func (c *nativeStoreClient) GetNative(collection, uuid, tid string) (nMsg *nativeMSG, found bool, err error) {
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
	req.Header.Set("Connection", "close")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("unsuccessful request for fetching native content uuid=%v %v", uuid, err)
	}
	defer niceClose(resp)
	if resp.StatusCode == http.StatusNotFound {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, false, fmt.Errorf("unexpected status while fetching native content uuid=%v collection=%v status=%v", uuid, collection, resp.StatusCode)
	}

	nMsg = new(nativeMSG)
	nMsg.body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read response body for uuid=%v %v", uuid, err)
	}
	nMsg.contentType = resp.Header.Get(contentTypeHeader)
	if nMsg.contentType == "" {
		nMsg.contentType = "application/json"
	}
	nMsg.originSystemID = resp.Header.Get(originSystemIDHeader)
	return nMsg, true, nil
}

func niceClose(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		logrus.Warnf("Couldn't close response body %v", err)
	}
}
