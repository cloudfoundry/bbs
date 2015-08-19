package bbs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
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
	KeepContainer        = true
	DeleteContainer      = false
)

//go:generate counterfeiter -o fake_bbs/fake_client.go . Client

type Client interface {
	Domains() ([]string, error)
	UpsertDomain(domain string, ttl time.Duration) error

	ActualLRPGroups(models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupsByProcessGuid(processGuid string) ([]*models.ActualLRPGroup, error)
	ActualLRPGroupByProcessGuidAndIndex(processGuid string, index int) (*models.ActualLRPGroup, error)

	ClaimActualLRP(processGuid string, index int, instanceKey *models.ActualLRPInstanceKey) error
	StartActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error
	CrashActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error
	FailActualLRP(key *models.ActualLRPKey, errorMessage string) error
	RemoveActualLRP(processGuid string, index int) error
	RetireActualLRP(key *models.ActualLRPKey) error

	EvacuateClaimedActualLRP(*models.ActualLRPKey, *models.ActualLRPInstanceKey) (bool, error)
	EvacuateRunningActualLRP(*models.ActualLRPKey, *models.ActualLRPInstanceKey, *models.ActualLRPNetInfo, uint64) (bool, error)
	EvacuateStoppedActualLRP(*models.ActualLRPKey, *models.ActualLRPInstanceKey) (bool, error)
	EvacuateCrashedActualLRP(*models.ActualLRPKey, *models.ActualLRPInstanceKey, string) (bool, error)
	RemoveEvacuatingActualLRP(*models.ActualLRPKey, *models.ActualLRPInstanceKey) error

	DesiredLRPs(models.DesiredLRPFilter) ([]*models.DesiredLRP, error)
	DesiredLRPByProcessGuid(processGuid string) (*models.DesiredLRP, error)

	// Public Task Methods
	Tasks() ([]*models.Task, error)
	TasksByDomain(domain string) ([]*models.Task, error)
	TasksByCellID(cellId string) ([]*models.Task, error)
	TaskByGuid(guid string) (*models.Task, error)
	DesireTask(guid, domain string, def *models.TaskDefinition) error
	CancelTask(taskGuid string) error
	FailTask(taskGuid, failureReason string) error
	CompleteTask(taskGuid, cellId string, failed bool, failureReason, result string) error
	ResolvingTask(taskGuid string) error
	ResolveTask(taskGuid string) error
	SubscribeToEvents() (events.EventSource, error)
	ConvergeTasks(kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration) error

	// Internal Task Methods
	StartTask(taskGuid string, cellID string) (bool, error)
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
	response := models.DomainsResponse{}
	err := c.doRequest(DomainsRoute, nil, nil, nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Domains, response.Error
}

func (c *client) UpsertDomain(domain string, ttl time.Duration) error {
	request := models.UpsertDomainRequest{
		Domain: domain,
		Ttl:    uint32(ttl.Seconds()),
	}
	response := models.UpsertDomainResponse{}
	err := c.doRequest(UpsertDomainRoute, nil, nil, &request, &response)
	if err != nil {
		return err
	}
	return response.Error
}

func (c *client) ActualLRPGroups(filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupsRequest{
		Domain: filter.Domain,
		CellId: filter.CellID,
	}
	response := models.ActualLRPGroupsResponse{}
	err := c.doRequest(ActualLRPGroupsRoute, nil, nil, &request, &response)
	if err != nil {
		return nil, err
	}

	return response.ActualLrpGroups, response.Error
}

func (c *client) ActualLRPGroupsByProcessGuid(processGuid string) ([]*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupsByProcessGuidRequest{
		ProcessGuid: processGuid,
	}
	response := models.ActualLRPGroupsResponse{}
	err := c.doRequest(ActualLRPGroupsByProcessGuidRoute, nil, nil, &request, &response)
	if err != nil {
		return nil, err
	}

	return response.ActualLrpGroups, response.Error
}

func (c *client) ActualLRPGroupByProcessGuidAndIndex(processGuid string, index int) (*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupByProcessGuidAndIndexRequest{
		ProcessGuid: processGuid,
		Index:       int32(index),
	}
	response := models.ActualLRPGroupResponse{}
	err := c.doRequest(ActualLRPGroupByProcessGuidAndIndexRoute, nil, nil, &request, &response)
	if err != nil {
		return nil, err
	}

	return response.ActualLrpGroup, response.Error
}

func (c *client) ClaimActualLRP(processGuid string, index int, instanceKey *models.ActualLRPInstanceKey) error {
	request := models.ClaimActualLRPRequest{
		ProcessGuid:          processGuid,
		Index:                int32(index),
		ActualLrpInstanceKey: instanceKey,
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(ClaimActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err
	}
	return response.Error
}

func (c *client) StartActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error {
	request := models.StartActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ActualLrpNetInfo:     netInfo,
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(StartActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err

	}
	return response.Error
}

func (c *client) CrashActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	request := models.CrashActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ErrorMessage:         errorMessage,
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(CrashActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err

	}
	return response.Error
}

func (c *client) FailActualLRP(key *models.ActualLRPKey, errorMessage string) error {
	request := models.FailActualLRPRequest{
		ActualLrpKey: key,
		ErrorMessage: errorMessage,
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(FailActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err

	}
	return response.Error
}

func (c *client) RetireActualLRP(key *models.ActualLRPKey) error {
	request := models.RetireActualLRPRequest{
		ActualLrpKey: key,
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(RetireActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err

	}
	return response.Error
}

func (c *client) RemoveActualLRP(processGuid string, index int) error {
	request := models.RemoveActualLRPRequest{
		ProcessGuid: processGuid,
		Index:       int32(index),
	}
	response := models.ActualLRPLifecycleResponse{}
	err := c.doRequest(RemoveActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err
	}
	return response.Error
}

func (c *client) EvacuateClaimedActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	return c.doEvacRequest(EvacuateClaimedActualLRPRoute, KeepContainer, &models.EvacuateClaimedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	})
}

func (c *client) EvacuateCrashedActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error) {
	return c.doEvacRequest(EvacuateCrashedActualLRPRoute, DeleteContainer, &models.EvacuateCrashedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ErrorMessage:         errorMessage,
	})
}

func (c *client) EvacuateStoppedActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	return c.doEvacRequest(EvacuateStoppedActualLRPRoute, DeleteContainer, &models.EvacuateStoppedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	})
}

func (c *client) EvacuateRunningActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, ttl uint64) (bool, error) {
	return c.doEvacRequest(EvacuateRunningActualLRPRoute, KeepContainer, &models.EvacuateRunningActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ActualLrpNetInfo:     netInfo,
		Ttl:                  ttl,
	})
}

func (c *client) RemoveEvacuatingActualLRP(key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	request := models.RemoveEvacuatingActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	}

	response := models.RemoveEvacuatingActualLRPResponse{}
	err := c.doRequest(RemoveEvacuatingActualLRPRoute, nil, nil, &request, &response)
	if err != nil {
		return err
	}

	return response.Error
}

func (c *client) DesiredLRPs(filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	request := models.DesiredLRPsRequest{
		Domain: filter.Domain,
	}
	response := models.DesiredLRPsResponse{}
	err := c.doRequest(DesiredLRPsRoute, nil, nil, &request, &response)
	if err != nil {
		return nil, err
	}

	return response.DesiredLrps, response.Error
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

func (c *client) DesireTask(taskGuid, domain string, taskDef *models.TaskDefinition) error {
	req := &models.DesireTaskRequest{
		TaskGuid:       taskGuid,
		Domain:         domain,
		TaskDefinition: taskDef,
	}
	return c.doRequest(DesireTaskRoute, nil, nil, req, nil)
}

func (c *client) StartTask(taskGuid string, cellId string) (bool, error) {
	req := &models.StartTaskRequest{
		TaskGuid: taskGuid,
		CellId:   cellId,
	}
	res := &models.StartTaskResponse{}
	err := c.doRequest(StartTaskRoute, nil, nil, req, res)
	return res.GetShouldStart(), err
}

func (c *client) CancelTask(taskGuid string) error {
	req := &models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	return c.doRequest(CancelTaskRoute, nil, nil, req, nil)
}

func (c *client) ResolvingTask(taskGuid string) error {
	req := &models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	return c.doRequest(ResolvingTaskRoute, nil, nil, req, nil)
}

func (c *client) ResolveTask(taskGuid string) error {
	req := &models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	return c.doRequest(ResolveTaskRoute, nil, nil, req, nil)
}

func (c *client) FailTask(taskGuid, failureReason string) error {
	req := &models.FailTaskRequest{
		TaskGuid:      taskGuid,
		FailureReason: failureReason,
	}
	return c.doRequest(FailTaskRoute, nil, nil, req, nil)
}

func (c *client) CompleteTask(taskGuid, cellId string, failed bool, failureReason, result string) error {
	req := &models.CompleteTaskRequest{
		TaskGuid:      taskGuid,
		CellId:        cellId,
		Failed:        failed,
		FailureReason: failureReason,
		Result:        result,
	}
	return c.doRequest(CompleteTaskRoute, nil, nil, req, nil)
}

func (c *client) ConvergeTasks(
	kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration,
) error {
	req := &models.ConvergeTasksRequest{
		KickTaskDuration:            kickTaskDuration.Nanoseconds(),
		ExpirePendingTaskDuration:   expirePendingTaskDuration.Nanoseconds(),
		ExpireCompletedTaskDuration: expireCompletedTaskDuration.Nanoseconds(),
	}
	return c.doRequest(ConvergeTasksRoute, nil, nil, req, nil)
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

func (c *client) doEvacRequest(route string, defaultKeepContainer bool, request proto.Message) (bool, error) {
	var response models.EvacuationResponse
	err := c.doRequest(route, nil, nil, request, &response)
	if err != nil {
		return defaultKeepContainer, err
	}

	return response.KeepContainer, response.Error.ToError()
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, requestBody, responseBody proto.Message) error {
	req, err := c.createRequest(requestName, params, queryParams, requestBody)
	if err != nil {
		return err
	}
	return c.do(req, responseBody)
}

func (c *client) do(req *http.Request, responseObject proto.Message) error {
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
		if res.StatusCode > 299 {
			return handleErrorResponse(res)
		} else {
			return handleProtoResponse(res, responseObject)
		}
	} else {
		return handleNonProtoResponse(res)
	}
}

func handleErrorResponse(res *http.Response) error {
	errResponse := &models.Error{}
	err := handleProtoResponse(res, errResponse)
	if err != nil {
		return &models.Error{Type: models.InvalidProtobufMessage, Message: err.Error()}
	}
	return errResponse
}

func handleProtoResponse(res *http.Response, responseObject proto.Message) error {
	if responseObject == nil {
		return &models.Error{Type: models.InvalidRequest, Message: "responseObject cannot be nil"}
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &models.Error{Type: models.InvalidResponse, Message: fmt.Sprint("failed to read body: ", err.Error())}
	}

	err = proto.Unmarshal(buf, responseObject)
	if err != nil {
		return &models.Error{Type: models.InvalidProtobufMessage, Message: fmt.Sprintf("failed to unmarshal proto", err.Error())}
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
