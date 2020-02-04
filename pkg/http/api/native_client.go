package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/sirupsen/logrus"
)

type NativeMSG struct {
	Body           []byte
	ContentType    string
	OriginSystemID string
}

type NativeStoreClientInterface interface {
	GetNative(collection, uuid, tid string) (nativeContent *NativeMSG, found bool, err error)
}

type nativeStoreClient struct {
	httpClient    *http.Client
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

func (c *nativeStoreClient) GetNative(collection, uuid, tid string) (nMsg *NativeMSG, found bool, err error) {
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

	nMsg = new(NativeMSG)
	nMsg.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read response body for uuid=%v %v", uuid, err)
	}
	nMsg.ContentType = resp.Header.Get(contentTypeHeader)
	if nMsg.ContentType == "" {
		nMsg.ContentType = "application/json"
	}
	nMsg.OriginSystemID = resp.Header.Get(originSystemIDHeader)
	return nMsg, true, nil
}

func niceClose(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		log.Warnf("Couldn't close response body %v", err)
	}
}
