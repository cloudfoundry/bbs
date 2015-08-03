package bbs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/gogo/protobuf/proto"
	"github.com/tedsuo/rata"
	"github.com/vito/go-sse/sse"
)

const (
	ContentTypeHeader    = "Content-Type"
	XCfRouterErrorHeader = "X-Cf-Routererror"
	ProtoContentType     = "application/x-protobuf"
)

//go:generate counterfeiter -o fake_bbs/fake_client.go . Client

type Client interface {
	Domains() ([]string, error)
	UpsertDomain(domain string, ttl time.Duration) error

	ActualLRPGroups(models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupsByProcessGuid(processGuid string) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupByProcessGuidAndIndex(processGuid string, index int) (*models.ActualLRPGroup, error)

	// ActualLRP Lifecycle
	ClaimActualLRP(processGuid string, index int, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRP, error)
	StartActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRP, error)
	CrashActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error
	FailActualLRP(key *models.ActualLRPKey, errorMessage string) error
	RemoveActualLRP(processGuid string, index int) error
	RetireActualLRP(key *models.ActualLRPKey) error

	DesiredLRPs(models.DesiredLRPFilter) ([]*models.DesiredLRP, error)
	DesiredLRPByProcessGuid(processGuid string) (*models.DesiredLRP, error)

	Tasks() ([]*models.Task, error)
	TasksByDomain(domain string) ([]*models.Task, error)
	TasksByCellID(cellId string) ([]*models.Task, error)
	TaskByGuid(guid string) (*models.Task, error)

	SubscribeToEvents() (events.EventSource, error)
}

func NewClient(url string) Client {
	return &client{
		httpClient:          cf_http.NewClient(),
		streamingHTTPClient: cf_http.NewStreamingClient(),

		reqGen: rata.NewRequestGenerator(url, Routes),
	}
}

type client struct {
	httpClient          *http.Client
	streamingHTTPClient *http.Client

	reqGen *rata.RequestGenerator
}

func (c *client) Domains() ([]string, error) {
	var domains models.Domains
	err := c.doRequest(DomainsRoute, nil, nil, nil, &domains)
	return domains.GetDomains(), err
}

func (c *client) UpsertDomain(domain string, ttl time.Duration) error {
	req, err := c.createRequest(UpsertDomainRoute, rata.Params{"domain": domain}, nil, nil)
	if err != nil {
		return err
	}

	if ttl != 0 {
		req.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(ttl.Seconds())))
	}

	return c.do(req, nil)
}

func (c *client) ActualLRPGroups(filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	var actualLRPGroups models.ActualLRPGroups
	query := url.Values{}
	if filter.Domain != "" {
		query.Set("domain", filter.Domain)
	}
	if filter.CellID != "" {
		query.Set("cell_id", filter.CellID)
	}
	err := c.doRequest(ActualLRPGroupsRoute, nil, query, nil, &actualLRPGroups)
	return actualLRPGroups.GetActualLrpGroups(), err
}

func (c *client) ActualLRPGroupsByProcessGuid(processGuid string) ([]*models.ActualLRPGroup, error) {
	var actualLRPGroups models.ActualLRPGroups
	err := c.doRequest(ActualLRPGroupsByProcessGuidRoute, rata.Params{"process_guid": processGuid}, nil, nil, &actualLRPGroups)
	return actualLRPGroups.GetActualLrpGroups(), err
}

func (c *client) ActualLRPGroupByProcessGuidAndIndex(processGuid string, index int) (*models.ActualLRPGroup, error) {
	var actualLRPGroup models.ActualLRPGroup
	err := c.doRequest(ActualLRPGroupByProcessGuidAndIndexRoute,
		rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)},
		nil, nil, &actualLRPGroup)
	return &actualLRPGroup, err
}

func (c *client) ClaimActualLRP(processGuid string, index int, instanceKey *models.ActualLRPInstanceKey) (*models.ActualLRP, error) {
	var actualLRP models.ActualLRP
	request := models.ClaimActualLRPRequest{
		ProcessGuid:          processGuid,
		Index:                int32(index),
		ActualLrpInstanceKey: instanceKey,
	}
	err := c.doRequest(ClaimActualLRPRoute, nil, nil, &request, &actualLRP)
	return &actualLRP, err
}

func (c *client) StartActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) (*models.ActualLRP, error) {
	var actualLRP models.ActualLRP
	request := models.StartActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ActualLrpNetInfo:     netInfo,
	}
	err := c.doRequest(StartActualLRPRoute, nil, nil, &request, &actualLRP)
	return &actualLRP, err
}

func (c *client) CrashActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	request := models.CrashActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ErrorMessage:         errorMessage,
	}
	return c.doRequest(CrashActualLRPRoute, nil, nil, &request, nil)
}

func (c *client) FailActualLRP(key *models.ActualLRPKey, errorMessage string) error {
	request := models.FailActualLRPRequest{
		ActualLrpKey: key,
		ErrorMessage: errorMessage,
	}
	return c.doRequest(FailActualLRPRoute, nil, nil, &request, nil)
}

func (c *client) RemoveActualLRP(processGuid string, index int) error {
	err := c.doRequest(RemoveActualLRPRoute,
		rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)},
		nil, nil, nil)
	return err
}

func (c *client) RetireActualLRP(key *models.ActualLRPKey) error {
	request := models.RetireActualLRPRequest{
		ActualLrpKey: key,
	}
	return c.doRequest(RetireActualLRPRoute, nil, nil, &request, nil)
}

func (c *client) DesiredLRPs(filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	var desiredLRPs models.DesiredLRPs
	query := url.Values{}
	if filter.Domain != "" {
		query.Set("domain", filter.Domain)
	}
	err := c.doRequest(DesiredLRPsRoute, nil, query, nil, &desiredLRPs)
	return desiredLRPs.GetDesiredLrps(), err
}

func (c *client) DesiredLRPByProcessGuid(processGuid string) (*models.DesiredLRP, error) {
	var desiredLRP models.DesiredLRP
	err := c.doRequest(DesiredLRPByProcessGuidRoute,
		rata.Params{"process_guid": processGuid},
		nil, nil, &desiredLRP)
	return &desiredLRP, err
}

func (c *client) Tasks() ([]*models.Task, error) {
	var tasks models.Tasks
	err := c.doRequest(TasksRoute, nil, nil, nil, &tasks)
	return tasks.Tasks, err
}

func (c *client) TasksByDomain(domain string) ([]*models.Task, error) {
	var tasks models.Tasks
	query := url.Values{}
	query.Set("domain", domain)
	err := c.doRequest(TasksRoute, nil, query, nil, &tasks)
	return tasks.Tasks, err
}

func (c *client) TasksByCellID(cellId string) ([]*models.Task, error) {
	var tasks models.Tasks
	query := url.Values{}
	query.Set("cell_id", cellId)
	err := c.doRequest(TasksRoute, nil, query, nil, &tasks)
	return tasks.Tasks, err
}

func (c *client) TaskByGuid(taskGuid string) (*models.Task, error) {
	var task models.Task
	err := c.doRequest(TaskByGuidRoute,
		rata.Params{"task_guid": taskGuid},
		nil, nil, &task)
	return &task, err
}

func (c *client) SubscribeToEvents() (events.EventSource, error) {
	eventSource, err := sse.Connect(c.streamingHTTPClient, time.Second, func() *http.Request {
		request, err := c.reqGen.CreateRequest(EventStreamRoute, nil, nil)
		if err != nil {
			panic(err) // totally shouldn't happen
		}

		return request
	})

	if err != nil {
		return nil, err
	}

	return events.NewEventSource(eventSource), nil
}

func (c *client) createRequest(requestName string, params rata.Params, queryParams url.Values, message proto.Message) (*http.Request, error) {
	var messageBody []byte
	var err error
	if message != nil {
		messageBody, err = proto.Marshal(message)
		if err != nil {
			return nil, err
		}
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(messageBody))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(messageBody))
	req.Header.Set("Content-Type", ProtoContentType)
	return req, nil
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, message proto.Message) error {
	req, err := c.createRequest(requestName, params, queryParams, request)
	if err != nil {
		return err
	}
	return c.do(req, message)
}

func (c *client) do(req *http.Request, responseObject interface{}) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var parsedContentType string
	if contentType, ok := res.Header[ContentTypeHeader]; ok {
		parsedContentType, _, _ = mime.ParseMediaType(contentType[0])
	}

	if routerError, ok := res.Header[XCfRouterErrorHeader]; ok {
		return &models.Error{Type: models.RouterError, Message: routerError[0]}
	}

	if parsedContentType == ProtoContentType {
		protoMessage, ok := responseObject.(proto.Message)
		if !ok && responseObject != nil {
			return &models.Error{Type: models.InvalidRequest, Message: "cannot read response body"}
		}
		return handleProtoResponse(res, protoMessage)
	} else {
		return handleNonProtoResponse(res)
	}
}

func handleProtoResponse(res *http.Response, responseObject proto.Message) error {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &models.Error{Type: models.InvalidResponse, Message: err.Error()}
	}

	if res.StatusCode > 299 {
		errResponse := &models.Error{}
		err = proto.Unmarshal(buf, errResponse)
		if err != nil {
			return &models.Error{Type: models.InvalidProtobufMessage, Message: err.Error()}
		}
		return errResponse
	}

	err = proto.Unmarshal(buf, responseObject)
	if err != nil {
		return &models.Error{Type: models.InvalidProtobufMessage, Message: err.Error()}
	}
	return nil
}

func handleNonProtoResponse(res *http.Response) error {
	if res.StatusCode > 299 {
		return &models.Error{
			Type:    models.InvalidResponse,
			Message: fmt.Sprintf("Invalid Response with status code: %d", res.StatusCode),
		}
	}
	return nil
}
