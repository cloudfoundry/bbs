package auctionhandlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter . Client
type Client interface {
	RequestLRPAuctions(lrpStart []*models.LRPStartRequest) error
	RequestTaskAuctions(tasks []oldmodels.Task) error
}

type auctioneerClient struct {
	httpClient *http.Client
	url        string
}

func NewClient(auctioneerURL string) Client {
	return &auctioneerClient{
		httpClient: cf_http.NewClient(),
		url:        auctioneerURL,
	}
}

func (c *auctioneerClient) RequestLRPAuctions(lrpStarts []*models.LRPStartRequest) error {
	reqGen := rata.NewRequestGenerator(c.url, Routes)

	payload, err := json.Marshal(lrpStarts)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateLRPAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (c *auctioneerClient) RequestTaskAuctions(tasks []oldmodels.Task) error {
	reqGen := rata.NewRequestGenerator(c.url, Routes)

	payload, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateTaskAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
