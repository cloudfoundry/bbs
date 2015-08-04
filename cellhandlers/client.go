package cellhandlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter . Client

type Client interface {
	StopLRPInstance(cellURL string, key models.ActualLRPKey, instanceKey models.ActualLRPInstanceKey) error
	CancelTask(cellURL string, taskGuid string) error
}

type client struct {
	httpClient *http.Client
}

func NewClient() Client {
	return &client{
		httpClient: cf_http.NewClient(),
	}
}

func (c *client) StopLRPInstance(
	cellURL string,
	key models.ActualLRPKey,
	instanceKey models.ActualLRPInstanceKey,
) error {
	reqGen := rata.NewRequestGenerator(cellURL, Routes)

	req, err := reqGen.CreateRequest(StopLRPInstanceRoute, stopParamsFromLRP(key, instanceKey), nil)
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

func (c *client) CancelTask(cellURL string, taskGuid string) error {
	reqGen := rata.NewRequestGenerator(cellURL, Routes)

	req, err := reqGen.CreateRequest(CancelTaskRoute, rata.Params{"task_guid": taskGuid}, nil)
	if err != nil {
		return err
	}

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

func stopParamsFromLRP(
	key models.ActualLRPKey,
	instanceKey models.ActualLRPInstanceKey,
) rata.Params {
	return rata.Params{
		"process_guid":  key.ProcessGuid,
		"instance_guid": instanceKey.InstanceGuid,
		"index":         strconv.Itoa(int(key.Index)),
	}
}
