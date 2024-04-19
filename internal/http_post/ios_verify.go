package http_post

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awa/go-iap/appstore"
	"io/ioutil"
	"net/http"
	"time"
)

var ErrAppStoreServer = errors.New("AppStore server error")

const (
	// SandboxURL is the endpoint for sandbox environment.
	SandboxURL string = "https://sandbox.itunes.apple.com/verifyReceipt"
	// ProductionURL is the endpoint for production environment.
	ProductionURL string = "https://buy.itunes.apple.com/verifyReceipt"
	// ContentType is the request content-type for apple store.
	ContentType string = "application/json; charset=utf-8"
)

type StatusResponse struct {
	Status    int    `json:"status"`
	Exception string `json:"exception"`
}

type Client struct {
	ProductionURL string
	SandboxURL    string
	httpCli       *http.Client
}

func New() *Client {
	client := &Client{
		ProductionURL: ProductionURL,
		SandboxURL:    SandboxURL,
		httpCli: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	return client
}

func (c *Client) verify(url string, reqBody appstore.IAPRequest, result interface{}) (code StatusResponse, err error) {
	var r StatusResponse
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(reqBody); err != nil {
		return r, err
	}

	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return r, err
	}
	req.Header.Set("Content-Type", ContentType)
	ctx := context.Background()
	req = req.WithContext(ctx)
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		r.Status = resp.StatusCode
		return r, fmt.Errorf("Received http status code %d from the App Store: %w", resp.StatusCode, ErrAppStoreServer)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(buf, &result)
	if err != nil {
		return r, err
	}
	// https://developer.apple.com/library/content/technotes/tn2413/_index.html#//apple_ref/doc/uid/DTS40016228-CH1-RECEIPTURL
	err = json.Unmarshal(buf, &r)
	if err != nil {
		return r, err
	}
	return r, nil

	//return c.parseResponse(resp, result, ctx, reqBody)
}

func Start() {
	c := New()
	var appstoreResponseNew = new(appstore.IAPResponse)
	params := appstore.IAPRequest{
		ReceiptData: "111111111",
	}

	status, err := c.verify(c.ProductionURL, params, appstoreResponseNew)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("status", status)
}
