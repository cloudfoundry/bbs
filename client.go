package bbs

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/trace"
	cfhttp "code.cloudfoundry.org/cfhttp/v2"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/tlsconfig"
	"github.com/tedsuo/rata"
	"github.com/vito/go-sse/sse"
	"google.golang.org/protobuf/proto"
)

const (
	ContentTypeHeader    = "Content-Type"
	XCfRouterErrorHeader = "X-Cf-Routererror"
	ProtoContentType     = "application/x-protobuf"
	KeepContainer        = true
	DeleteContainer      = false
	DefaultRetryCount    = 3

	InvalidResponseMessage = "Invalid Response with status code: %d"
)

var EndpointNotFoundErr = models.NewError(models.Error_InvalidResponse, fmt.Sprintf(InvalidResponseMessage, 404))

//go:generate counterfeiter -generate

//counterfeiter:generate -o fake_bbs/fake_internal_client.go . InternalClient
//counterfeiter:generate -o fake_bbs/fake_client.go . Client

/*
The InternalClient interface exposes all available endpoints of the BBS server,
including private endpoints which should be used exclusively by internal Diego
components. To interact with the BBS from outside of Diego, the Client
should be used instead.
*/
type InternalClient interface {
	Client

	ClaimActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error
	StartActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, internalRoutes []*models.ActualLRPInternalRoute, metricTags map[string]string, routable bool, availabilityZone string) error
	CrashActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error
	FailActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, errorMessage string) error
	RemoveActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error

	EvacuateClaimedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error)
	EvacuateRunningActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo, internalRoutes []*models.ActualLRPInternalRoute, metricTags map[string]string, routable bool, availabilityZone string) (bool, error)
	EvacuateStoppedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error)
	EvacuateCrashedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error)
	RemoveEvacuatingActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error

	StartTask(logger lager.Logger, traceID string, taskGuid string, cellID string) (bool, error)
	FailTask(logger lager.Logger, traceID string, taskGuid, failureReason string) error
	RejectTask(logger lager.Logger, traceID string, taskGuid, failureReason string) error
	CompleteTask(logger lager.Logger, traceID string, taskGuid, cellId string, failed bool, failureReason, result string) error
}

/*
The External InternalClient can be used to access the BBS's public functionality.
It exposes methods for basic LRP and Task Lifecycles, Domain manipulation, and
event subscription.
*/
type Client interface {
	ExternalTaskClient
	ExternalDomainClient
	ExternalActualLRPClient
	ExternalDesiredLRPClient
	ExternalEventClient

	// Returns true if the BBS server is reachable
	Ping(logger lager.Logger, traceID string) bool

	// Lists all Cells
	Cells(logger lager.Logger, traceID string) ([]*models.CellPresence, error)
}

/*
The ExternalTaskClient is used to access Diego's ability to run one-off tasks.
More information about this API can be found in the bbs docs:

https://code.cloudfoundry.org/bbs/tree/master/doc/tasks.md
*/
type ExternalTaskClient interface {
	// Creates a Task from the given TaskDefinition
	DesireTask(logger lager.Logger, traceID string, guid string, domain string, def *models.TaskDefinition) error

	// Lists all Tasks
	Tasks(logger lager.Logger, traceID string) ([]*models.Task, error)

	// List all Tasks that match filter
	TasksWithFilter(logger lager.Logger, traceID string, filter models.TaskFilter) ([]*models.Task, error)

	// Lists all Tasks of the given domain
	TasksByDomain(logger lager.Logger, traceID string, domain string) ([]*models.Task, error)

	// Lists all Tasks on the given cell
	TasksByCellID(logger lager.Logger, traceID string, cellId string) ([]*models.Task, error)

	// Returns the Task with the given guid
	TaskByGuid(logger lager.Logger, traceID string, guid string) (*models.Task, error)

	// Cancels the Task with the given task guid
	CancelTask(logger lager.Logger, traceID string, taskGuid string) error

	// Resolves a Task with the given guid
	ResolvingTask(logger lager.Logger, traceID string, taskGuid string) error

	// Deletes a completed task with the given guid
	DeleteTask(logger lager.Logger, traceID string, taskGuid string) error
}

/*
The ExternalDomainClient is used to access and update Diego's domains.
*/
type ExternalDomainClient interface {
	// Lists the active domains
	Domains(logger lager.Logger, traceID string) ([]string, error)

	// Creates a domain or bumps the ttl on an existing domain
	UpsertDomain(logger lager.Logger, traceID string, domain string, ttl time.Duration) error
}

/*
The ExternalActualLRPClient is used to access and retire Actual LRPs
*/
type ExternalActualLRPClient interface {
	// Returns all ActualLRPs matching the given ActualLRPFilter
	ActualLRPs(lager.Logger, string, models.ActualLRPFilter) ([]*models.ActualLRP, error)

	// Returns all ActualLRPGroups matching the given ActualLRPFilter
	//lint:ignore SA1019 - deprecated function returning deprecated data
	// Deprecated: use ActualLRPs instead
	ActualLRPGroups(lager.Logger, string, models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)

	// Returns all ActualLRPGroups that have the given process guid
	//lint:ignore SA1019 - deprecated function returning deprecated data
	// Deprecated: use ActualLRPs instead
	ActualLRPGroupsByProcessGuid(logger lager.Logger, traceID string, processGuid string) ([]*models.ActualLRPGroup, error)

	// Returns the ActualLRPGroup with the given process guid and instance index
	//lint:ignore SA1019 - deprecated function returning deprecated data
	// Deprecated: use ActualLRPs instead
	ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, traceID string, processGuid string, index int) (*models.ActualLRPGroup, error)

	// Shuts down the ActualLRP matching the given ActualLRPKey, but does not modify the desired state
	RetireActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey) error
}

/*
The ExternalDesiredLRPClient is used to access and manipulate Desired LRPs.
*/
type ExternalDesiredLRPClient interface {
	// Lists all DesiredLRPs that match the given DesiredLRPFilter
	DesiredLRPs(lager.Logger, string, models.DesiredLRPFilter) ([]*models.DesiredLRP, error)

	// Returns the DesiredLRP with the given process guid
	DesiredLRPByProcessGuid(logger lager.Logger, traceID string, processGuid string) (*models.DesiredLRP, error)

	// Returns all DesiredLRPSchedulingInfos that match the given DesiredLRPFilter
	DesiredLRPSchedulingInfos(lager.Logger, string, models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error)

	//Returns the DesiredLRPSchedulingInfo that matches the given process guid
	DesiredLRPSchedulingInfoByProcessGuid(logger lager.Logger, traceID string, processGuid string) (*models.DesiredLRPSchedulingInfo, error)

	// Returns all DesiredLRPRoutingInfos that match the given DesiredLRPFilter
	DesiredLRPRoutingInfos(lager.Logger, string, models.DesiredLRPFilter) ([]*models.DesiredLRP, error)

	// Creates the given DesiredLRP and its corresponding ActualLRPs
	DesireLRP(lager.Logger, string, *models.DesiredLRP) error

	// Updates the DesiredLRP matching the given process guid
	UpdateDesiredLRP(logger lager.Logger, traceID string, processGuid string, update *models.DesiredLRPUpdate) error

	// Removes the DesiredLRP matching the given process guid
	RemoveDesiredLRP(logger lager.Logger, traceID string, processGuid string) error
}

/*
The ExternalEventClient is used to subscribe to groups of Events.
*/
type ExternalEventClient interface {
	// Deprecated: use SubscribeToInstanceEvents instead
	SubscribeToEvents(logger lager.Logger) (events.EventSource, error)

	SubscribeToInstanceEvents(logger lager.Logger) (events.EventSource, error)
	SubscribeToTaskEvents(logger lager.Logger) (events.EventSource, error)

	// Deprecated: use SubscribeToInstanceEventsByCellID instead
	SubscribeToEventsByCellID(logger lager.Logger, cellId string) (events.EventSource, error)

	SubscribeToInstanceEventsByCellID(logger lager.Logger, cellId string) (events.EventSource, error)
}

type ClientConfig struct {
	URL                    string
	IsTLS                  bool
	CAFile                 string
	CertFile               string
	KeyFile                string
	ClientSessionCacheSize int
	MaxIdleConnsPerHost    int
	InsecureSkipVerify     bool
	Retries                int
	RetryInterval          time.Duration // Only affects streaming client, not the http client
	RequestTimeout         time.Duration // Only affects the http client, not the streaming client
}

func NewClient(url, caFile, certFile, keyFile string, clientSessionCacheSize, maxIdleConnsPerHost int) (InternalClient, error) {
	return NewClientWithConfig(ClientConfig{
		URL:                    url,
		IsTLS:                  true,
		CAFile:                 caFile,
		CertFile:               certFile,
		KeyFile:                keyFile,
		ClientSessionCacheSize: clientSessionCacheSize,
		MaxIdleConnsPerHost:    maxIdleConnsPerHost,
	})
}

func NewSecureSkipVerifyClient(url, certFile, keyFile string, clientSessionCacheSize, maxIdleConnsPerHost int) (InternalClient, error) {
	return NewClientWithConfig(ClientConfig{
		URL:                    url,
		IsTLS:                  true,
		CAFile:                 "",
		CertFile:               certFile,
		KeyFile:                keyFile,
		ClientSessionCacheSize: clientSessionCacheSize,
		MaxIdleConnsPerHost:    maxIdleConnsPerHost,
		InsecureSkipVerify:     true,
	})
}

func NewClientWithConfig(cfg ClientConfig) (InternalClient, error) {
	if cfg.Retries == 0 {
		cfg.Retries = DefaultRetryCount
	}

	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = time.Second
	}

	if cfg.InsecureSkipVerify {
		cfg.CAFile = ""
	}

	if cfg.IsTLS {
		return newSecureClient(cfg)
	} else {
		return newClient(cfg), nil
	}
}

func newClient(cfg ClientConfig) *client {
	return &client{
		httpClient:          cfhttp.NewClient(cfhttp.WithRequestTimeout(cfg.RequestTimeout)),
		streamingHTTPClient: cfhttp.NewClient(cfhttp.WithStreamingDefaults()),
		reqGen:              rata.NewRequestGenerator(cfg.URL, Routes),
		requestRetryCount:   cfg.Retries,
		retryInterval:       cfg.RetryInterval,
	}
}
func newSecureClient(cfg ClientConfig) (InternalClient, error) {
	bbsURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	if bbsURL.Scheme != "https" {
		return nil, errors.New("Expected https URL")
	}

	var clientOpts []tlsconfig.ClientOption
	if !cfg.InsecureSkipVerify {
		clientOpts = append(clientOpts, tlsconfig.WithAuthorityFromFile(cfg.CAFile))
	}

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(cfg.CertFile, cfg.KeyFile),
	).Client(clientOpts...)
	if err != nil {
		return nil, err
	}
	tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(cfg.ClientSessionCacheSize)

	tlsConfig.InsecureSkipVerify = cfg.InsecureSkipVerify

	httpClient := cfhttp.NewClient(
		cfhttp.WithRequestTimeout(cfg.RequestTimeout),
		cfhttp.WithTLSConfig(tlsConfig),
		cfhttp.WithMaxIdleConnsPerHost(cfg.MaxIdleConnsPerHost),
	)
	streamingClient := cfhttp.NewClient(
		cfhttp.WithStreamingDefaults(),
		cfhttp.WithTLSConfig(tlsConfig),
		cfhttp.WithMaxIdleConnsPerHost(cfg.MaxIdleConnsPerHost),
	)

	return &client{
		httpClient:          httpClient,
		streamingHTTPClient: streamingClient,
		reqGen:              rata.NewRequestGenerator(cfg.URL, Routes),
		requestRetryCount:   cfg.Retries,
		retryInterval:       cfg.RetryInterval,
	}, nil
}

type client struct {
	httpClient          *http.Client
	streamingHTTPClient *http.Client
	reqGen              *rata.RequestGenerator
	requestRetryCount   int
	retryInterval       time.Duration
}

func (c *client) Ping(logger lager.Logger, traceID string) bool {
	protoResponse := models.ProtoPingResponse{}
	err := c.doRequest(logger, traceID, PingRoute_r0, nil, nil, nil, &protoResponse)
	if err != nil {
		return false
	}
	return protoResponse.Available
}

func (c *client) Domains(logger lager.Logger, traceID string) ([]string, error) {
	protoResponse := models.ProtoDomainsResponse{}
	err := c.doRequest(logger, traceID, DomainsRoute_r0, nil, nil, nil, &protoResponse)
	if err != nil {
		return nil, err
	}
	response := protoResponse.FromProto()
	return response.Domains, response.Error.ToError()
}

func (c *client) UpsertDomain(logger lager.Logger, traceID string, domain string, ttl time.Duration) error {
	request := models.UpsertDomainRequest{
		Domain: domain,
		Ttl:    uint32(ttl.Seconds()),
	}
	protoResponse := models.ProtoUpsertDomainResponse{}
	err := c.doRequest(logger, traceID, UpsertDomainRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err
	}
	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) ActualLRPs(logger lager.Logger, traceID string, filter models.ActualLRPFilter) ([]*models.ActualLRP, error) {
	request := models.ActualLRPsRequest{
		Domain:      filter.Domain,
		CellId:      filter.CellID,
		ProcessGuid: filter.ProcessGuid,
	}
	if filter.Index != nil {
		request.SetIndex(filter.Index)
	}
	protoResponse := models.ProtoActualLRPsResponse{}
	err := c.doRequest(logger, traceID, ActualLRPsRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.ActualLrps, response.Error.ToError()
}

// Deprecated: use ActualLRPs instead
func (c *client) ActualLRPGroups(logger lager.Logger, traceID string, filter models.ActualLRPFilter) ([]*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupsRequest{
		Domain: filter.Domain,
		CellId: filter.CellID,
	}
	protoResponse := models.ProtoActualLRPGroupsResponse{}
	err := c.doRequest(logger, traceID, ActualLRPGroupsRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.ActualLrpGroups, response.Error.ToError()
}

// Deprecated: use ActaulLRPs instead
func (c *client) ActualLRPGroupsByProcessGuid(logger lager.Logger, traceID string, processGuid string) ([]*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupsByProcessGuidRequest{
		ProcessGuid: processGuid,
	}
	protoResponse := models.ProtoActualLRPGroupsResponse{}
	err := c.doRequest(logger, traceID, ActualLRPGroupsByProcessGuidRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.ActualLrpGroups, response.Error.ToError()
}

// Deprecated: use ActaulLRPs instead
func (c *client) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, traceID string, processGuid string, index int) (*models.ActualLRPGroup, error) {
	request := models.ActualLRPGroupByProcessGuidAndIndexRequest{
		ProcessGuid: processGuid,
		Index:       int32(index),
	}
	protoResponse := models.ProtoActualLRPGroupResponse{}
	err := c.doRequest(logger, traceID, ActualLRPGroupByProcessGuidAndIndexRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.ActualLrpGroup, response.Error.ToError()
}

func (c *client) ClaimActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	fmt.Println("BBS CLIENT: ClaimActualLRP")
	request := models.ClaimActualLRPRequest{
		ProcessGuid:          key.ProcessGuid,
		Index:                key.Index,
		ActualLrpInstanceKey: instanceKey,
	}
	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	err := c.doRequest(logger, traceID, ClaimActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) StartActualLRP(logger lager.Logger,
	traceID string,
	key *models.ActualLRPKey,
	instanceKey *models.ActualLRPInstanceKey,
	netInfo *models.ActualLRPNetInfo,
	internalRoutes []*models.ActualLRPInternalRoute,
	metricTags map[string]string,
	routable bool,
	availabilityZone string,
) error {
	fmt.Println("BBS CLIENT: StartActualLRP")
	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	request := &models.StartActualLRPRequest{
		ActualLrpKey:            key,
		ActualLrpInstanceKey:    instanceKey,
		ActualLrpNetInfo:        netInfo,
		ActualLrpInternalRoutes: internalRoutes,
		MetricTags:              metricTags,
		AvailabilityZone:        availabilityZone,
	}
	request.SetRoutable(&routable)
	err := c.doRequest(logger, traceID, StartActualLRPRoute_r1, nil, nil, request.ToProto(), &protoResponse)
	if err != nil && err == EndpointNotFoundErr {
		startActualLrpRequest := &models.StartActualLRPRequest{
			ActualLrpKey:         key,
			ActualLrpInstanceKey: instanceKey,
			ActualLrpNetInfo:     netInfo,
		}
		err = c.doRequest(logger, traceID, StartActualLRPRoute_r0, nil, nil, startActualLrpRequest.ToProto(), &protoResponse)
	}

	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) CrashActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error {
	request := models.CrashActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ErrorMessage:         errorMessage,
	}
	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	err := c.doRequest(logger, traceID, CrashActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err

	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) FailActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, errorMessage string) error {
	request := models.FailActualLRPRequest{
		ActualLrpKey: key,
		ErrorMessage: errorMessage,
	}
	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	err := c.doRequest(logger, traceID, FailActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err

	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) RetireActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey) error {
	request := models.RetireActualLRPRequest{
		ActualLrpKey: key,
	}
	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	err := c.doRequest(logger, traceID, RetireActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err

	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) RemoveActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	request := models.RemoveActualLRPRequest{
		ProcessGuid:          key.ProcessGuid,
		Index:                key.Index,
		ActualLrpInstanceKey: instanceKey,
	}

	protoResponse := models.ProtoActualLRPLifecycleResponse{}
	err := c.doRequest(logger, traceID, RemoveActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) EvacuateClaimedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	request := &models.EvacuateClaimedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	}
	return c.doEvacRequest(logger, traceID, EvacuateClaimedActualLRPRoute_r0, KeepContainer, request.ToProto())
}

func (c *client) EvacuateCrashedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error) {
	request := &models.EvacuateCrashedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
		ErrorMessage:         errorMessage,
	}
	return c.doEvacRequest(logger, traceID, EvacuateCrashedActualLRPRoute_r0, DeleteContainer, request.ToProto())
}

func (c *client) EvacuateStoppedActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error) {
	request := &models.EvacuateStoppedActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	}
	return c.doEvacRequest(logger, traceID, EvacuateStoppedActualLRPRoute_r0, DeleteContainer, request.ToProto())
}

func (c *client) EvacuateRunningActualLRP(logger lager.Logger,
	traceID string,
	key *models.ActualLRPKey,
	instanceKey *models.ActualLRPInstanceKey,
	netInfo *models.ActualLRPNetInfo,
	internalRoutes []*models.ActualLRPInternalRoute,
	metricTags map[string]string,
	routable bool,
	availabilityZone string,
) (bool, error) {
	request := &models.EvacuateRunningActualLRPRequest{
		ActualLrpKey:            key,
		ActualLrpInstanceKey:    instanceKey,
		ActualLrpNetInfo:        netInfo,
		ActualLrpInternalRoutes: internalRoutes,
		MetricTags:              metricTags,
		AvailabilityZone:        availabilityZone,
	}
	request.SetRoutable(&routable)
	keepContainer, err := c.doEvacRequest(logger, traceID, EvacuateRunningActualLRPRoute_r1, KeepContainer, request.ToProto())
	if err != nil && err == EndpointNotFoundErr {
		evacRunningActualLrpRequest := &models.EvacuateRunningActualLRPRequest{
			ActualLrpKey:         key,
			ActualLrpInstanceKey: instanceKey,
			ActualLrpNetInfo:     netInfo,
		}
		keepContainer, err = c.doEvacRequest(logger, traceID, EvacuateRunningActualLRPRoute_r0, KeepContainer, evacRunningActualLrpRequest.ToProto())
	}

	return keepContainer, err
}

func (c *client) RemoveEvacuatingActualLRP(logger lager.Logger, traceID string, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error {
	request := models.RemoveEvacuatingActualLRPRequest{
		ActualLrpKey:         key,
		ActualLrpInstanceKey: instanceKey,
	}

	protoResponse := models.ProtoRemoveEvacuatingActualLRPResponse{}
	err := c.doRequest(logger, traceID, RemoveEvacuatingActualLRPRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) DesiredLRPs(logger lager.Logger, traceID string, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	request := models.DesiredLRPsRequest(filter)
	protoResponse := models.ProtoDesiredLRPsResponse{}
	err := c.doRequest(logger, traceID, DesiredLRPsRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.DesiredLrps, response.Error.ToError()
}

func (c *client) DesiredLRPByProcessGuid(logger lager.Logger, traceID string, processGuid string) (*models.DesiredLRP, error) {
	request := models.DesiredLRPByProcessGuidRequest{
		ProcessGuid: processGuid,
	}
	protoResponse := models.ProtoDesiredLRPResponse{}
	err := c.doRequest(logger, traceID, DesiredLRPByProcessGuidRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.DesiredLrp, response.Error.ToError()
}

func (c *client) DesiredLRPSchedulingInfos(logger lager.Logger, traceID string, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error) {
	request := models.DesiredLRPsRequest(filter)
	protoResponse := models.ProtoDesiredLRPSchedulingInfosResponse{}
	err := c.doRequest(logger, traceID, DesiredLRPSchedulingInfosRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.DesiredLrpSchedulingInfos, response.Error.ToError()
}

func (c *client) DesiredLRPSchedulingInfoByProcessGuid(logger lager.Logger, traceID string, processGuid string) (*models.DesiredLRPSchedulingInfo, error) {
	request := models.DesiredLRPByProcessGuidRequest{
		ProcessGuid: processGuid,
	}
	protoResponse := models.ProtoDesiredLRPSchedulingInfoByProcessGuidResponse{}
	err := c.doRequest(logger, traceID, DesiredLRPSchedulingInfoByProcessGuid_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.DesiredLrpSchedulingInfo, response.Error.ToError()
}

func (c *client) DesiredLRPRoutingInfos(logger lager.Logger, traceID string, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error) {
	request := models.DesiredLRPsRequest(filter)
	protoResponse := models.ProtoDesiredLRPsResponse{}
	err := c.doRequest(logger, traceID, DesiredLRPRoutingInfosRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.DesiredLrps, response.Error.ToError()
}

func (c *client) doDesiredLRPLifecycleRequest(logger lager.Logger, traceID string, route string, request proto.Message) error {
	fmt.Println("doDesiredLRPLifecycleRequest")
	fmt.Printf("request: %+v\n", request)
	protoResponse := models.ProtoDesiredLRPLifecycleResponse{}
	fmt.Printf("protoResponse.FromProto() (empty): %+v\n", protoResponse.FromProto())
	err := c.doRequest(logger, traceID, route, nil, nil, request, &protoResponse)
	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	fmt.Printf("response: %+v\n", response)
	return response.Error.ToError()
}

func (c *client) DesireLRP(logger lager.Logger, traceID string, desiredLRP *models.DesiredLRP) error {
	fmt.Println("BBS CLIENT: DesireLRP")
	request := models.DesireLRPRequest{
		DesiredLrp: desiredLRP,
	}
	fmt.Printf("request: %+v\n", request)
	fmt.Printf("request.ToProto(): %+v\n", request.ToProto())
	return c.doDesiredLRPLifecycleRequest(logger, traceID, DesireDesiredLRPRoute_r2, request.ToProto())
}

func (c *client) UpdateDesiredLRP(logger lager.Logger, traceID string, processGuid string, update *models.DesiredLRPUpdate) error {
	request := models.UpdateDesiredLRPRequest{
		ProcessGuid: processGuid,
		Update:      update,
	}
	return c.doDesiredLRPLifecycleRequest(logger, traceID, UpdateDesiredLRPRoute_r0, request.ToProto())
}

func (c *client) RemoveDesiredLRP(logger lager.Logger, traceID string, processGuid string) error {
	request := models.RemoveDesiredLRPRequest{
		ProcessGuid: processGuid,
	}
	return c.doDesiredLRPLifecycleRequest(logger, traceID, RemoveDesiredLRPRoute_r0, request.ToProto())
}

func (c *client) Tasks(logger lager.Logger, traceID string) ([]*models.Task, error) {
	request := models.TasksRequest{}
	protoResponse := models.ProtoTasksResponse{}
	err := c.doRequest(logger, traceID, TasksRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Tasks, response.Error.ToError()
}

func (c *client) TasksWithFilter(logger lager.Logger, traceID string, filter models.TaskFilter) ([]*models.Task, error) {
	request := models.TasksRequest{
		Domain: filter.Domain,
		CellId: filter.CellID,
	}
	protoResponse := models.ProtoTasksResponse{}
	err := c.doRequest(logger, traceID, TasksRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Tasks, response.Error.ToError()
}

func (c *client) TasksByDomain(logger lager.Logger, traceID string, domain string) ([]*models.Task, error) {
	request := models.TasksRequest{
		Domain: domain,
	}
	protoResponse := models.ProtoTasksResponse{}
	err := c.doRequest(logger, traceID, TasksRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Tasks, response.Error.ToError()
}

func (c *client) TasksByCellID(logger lager.Logger, traceID string, cellId string) ([]*models.Task, error) {
	request := models.TasksRequest{
		CellId: cellId,
	}
	protoResponse := models.ProtoTasksResponse{}
	err := c.doRequest(logger, traceID, TasksRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Tasks, response.Error.ToError()
}

func (c *client) TaskByGuid(logger lager.Logger, traceID string, taskGuid string) (*models.Task, error) {
	request := models.TaskByGuidRequest{
		TaskGuid: taskGuid,
	}
	protoResponse := models.ProtoTaskResponse{}
	err := c.doRequest(logger, traceID, TaskByGuidRoute_r3, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Task, response.Error.ToError()
}

func (c *client) doTaskLifecycleRequest(logger lager.Logger, traceID string, route string, request proto.Message) error {
	protoResponse := models.ProtoTaskLifecycleResponse{}
	err := c.doRequest(logger, traceID, route, nil, nil, request, &protoResponse)
	if err != nil {
		return err
	}

	response := protoResponse.FromProto()
	return response.Error.ToError()
}

func (c *client) DesireTask(logger lager.Logger, traceID string, taskGuid, domain string, taskDef *models.TaskDefinition) error {
	route := DesireTaskRoute_r2
	request := models.DesireTaskRequest{
		TaskGuid:       taskGuid,
		Domain:         domain,
		TaskDefinition: taskDef,
	}
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) StartTask(logger lager.Logger, traceID string, taskGuid string, cellId string) (bool, error) {
	request := &models.StartTaskRequest{
		TaskGuid: taskGuid,
		CellId:   cellId,
	}
	protoResponse := models.ProtoStartTaskResponse{}
	err := c.doRequest(logger, traceID, StartTaskRoute_r0, nil, nil, request.ToProto(), &protoResponse)
	if err != nil {
		return false, err
	}

	response := protoResponse.FromProto()
	return response.ShouldStart, response.Error.ToError()
}

func (c *client) CancelTask(logger lager.Logger, traceID string, taskGuid string) error {
	request := models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	route := CancelTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) ResolvingTask(logger lager.Logger, traceID string, taskGuid string) error {
	request := models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	route := ResolvingTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) DeleteTask(logger lager.Logger, traceID string, taskGuid string) error {
	request := models.TaskGuidRequest{
		TaskGuid: taskGuid,
	}
	route := DeleteTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

// Deprecated: use CancelTask instead
func (c *client) FailTask(logger lager.Logger, traceID string, taskGuid string, failureReason string) error {
	request := models.FailTaskRequest{
		TaskGuid:      taskGuid,
		FailureReason: failureReason,
	}
	route := FailTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) RejectTask(logger lager.Logger, traceID string, taskGuid string, rejectionReason string) error {
	request := models.RejectTaskRequest{
		TaskGuid:        taskGuid,
		RejectionReason: rejectionReason,
	}
	route := RejectTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) CompleteTask(logger lager.Logger, traceID string, taskGuid string, cellId string, failed bool, failureReason, result string) error {
	request := models.CompleteTaskRequest{
		TaskGuid:      taskGuid,
		CellId:        cellId,
		Failed:        failed,
		FailureReason: failureReason,
		Result:        result,
	}
	route := CompleteTaskRoute_r0
	return c.doTaskLifecycleRequest(logger, traceID, route, request.ToProto())
}

func (c *client) subscribeToEvents(route string, cellId string) (events.EventSource, error) {
	request := models.EventsByCellId{
		CellId: cellId,
	}
	messageBody, err := proto.Marshal(request.ToProto())
	if err != nil {
		return nil, err
	}

	sseConfig := sse.Config{
		Client: c.streamingHTTPClient,
		RetryParams: sse.RetryParams{
			RetryInterval: c.retryInterval,
			MaxRetries:    uint16(c.requestRetryCount),
		},
		RequestCreator: func() *http.Request {
			request, err := c.reqGen.CreateRequest(route, nil, bytes.NewReader(messageBody))
			if err != nil {
				panic(err) // totally shouldn't happen
			}

			return request
		},
	}

	eventSource, err := sseConfig.Connect()
	if err != nil {
		return nil, err
	}

	return events.NewEventSource(eventSource), nil
}

// Deprecated: use SubscribeToInstanceEvents instead
func (c *client) SubscribeToEvents(logger lager.Logger) (events.EventSource, error) {
	return c.subscribeToEvents(LRPGroupEventStreamRoute_r1, "")
}

func (c *client) SubscribeToInstanceEvents(logger lager.Logger) (events.EventSource, error) {
	return c.subscribeToEvents(LRPInstanceEventStreamRoute_r1, "")
}

func (c *client) SubscribeToTaskEvents(logger lager.Logger) (events.EventSource, error) {
	return c.subscribeToEvents(TaskEventStreamRoute_r1, "")
}

// Deprecated: use SubscribeToInstanceEventsByCellID instead
func (c *client) SubscribeToEventsByCellID(logger lager.Logger, cellId string) (events.EventSource, error) {
	return c.subscribeToEvents(LRPGroupEventStreamRoute_r1, cellId)
}

func (c *client) SubscribeToInstanceEventsByCellID(logger lager.Logger, cellId string) (events.EventSource, error) {
	return c.subscribeToEvents(LRPInstanceEventStreamRoute_r1, cellId)
}

func (c *client) Cells(logger lager.Logger, traceID string) ([]*models.CellPresence, error) {
	protoResponse := models.ProtoCellsResponse{}
	err := c.doRequest(logger, traceID, CellsRoute_r0, nil, nil, nil, &protoResponse)
	if err != nil {
		return nil, err
	}

	response := protoResponse.FromProto()
	return response.Cells, response.Error.ToError()
}

func (c *client) createRequest(traceID string, requestName string, params rata.Params, queryParams url.Values, message proto.Message) (*http.Request, error) {
	fmt.Println("createRequest")
	var messageBody []byte
	var err error
	if message != nil {
		messageBody, err = proto.Marshal(message)
		fmt.Printf("messageBody: %+v\n", messageBody)
		if err != nil {
			return nil, err
		}
	}

	request, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(messageBody))
	fmt.Printf("request: %+v\n", request)
	if err != nil {
		return nil, err
	}

	request.URL.RawQuery = queryParams.Encode()
	request.ContentLength = int64(len(messageBody))
	request.Header.Set("Content-Type", ProtoContentType)
	request.Header.Set(trace.RequestIdHeader, traceID)
	return request, nil
}

func (c *client) doEvacRequest(logger lager.Logger, traceID string, route string, defaultKeepContainer bool, request proto.Message) (bool, error) {
	protoResponse := models.ProtoEvacuationResponse{}
	err := c.doRequest(logger, traceID, route, nil, nil, request, &protoResponse)
	if err != nil {
		return defaultKeepContainer, err
	}

	response := protoResponse.FromProto()
	return response.KeepContainer, response.Error.ToError()
}

func (c *client) doRequest(logger lager.Logger, traceID string, requestName string, params rata.Params, queryParams url.Values, requestBody, responseBody proto.Message) error {
	fmt.Println("doRequest")
	logger = logger.Session("do-request")
	var err error
	var request *http.Request

	for attempts := 0; attempts < c.requestRetryCount; attempts++ {
		logger.Debug("creating-request", lager.Data{"attempt": attempts + 1, "request_name": requestName})
		fmt.Printf("requestBody: %+v\n", requestBody)
		request, err = c.createRequest(traceID, requestName, params, queryParams, requestBody)
		fmt.Printf("request: %+v\n", request)
		if err != nil {
			logger.Error("failed-creating-request", err)
			return err
		}

		logger.Debug("doing-request", lager.Data{"attempt": attempts + 1, "request_path": request.URL.Path})

		start := time.Now().UnixNano()
		fmt.Printf("responseBody (empty): %+v\n", responseBody)
		err = c.do(request, responseBody)
		fmt.Printf("responseBody (filled): %+v\n", responseBody)
		finish := time.Now().UnixNano()

		if err != nil {
			logger.Error("failed-doing-request", err)
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					err = models.NewError(models.Error_Timeout, err.Error())
				}
			}
			time.Sleep(500 * time.Millisecond)
		} else {
			logger.Debug("complete", lager.Data{"request_path": request.URL.Path, "duration_in_ns": finish - start})
			break
		}
	}
	return err
}

func (c *client) do(request *http.Request, responseObject proto.Message) error {
	fmt.Println("do")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		// don't worry about errors when closing the body
		_ = response.Body.Close()
	}()

	var parsedContentType string
	if contentType, ok := response.Header[ContentTypeHeader]; ok {
		parsedContentType, _, _ = mime.ParseMediaType(contentType[0])
	}

	if routerError, ok := response.Header[XCfRouterErrorHeader]; ok {
		return models.NewError(models.Error_RouterError, routerError[0])
	}

	if parsedContentType == ProtoContentType {
		return handleProtoResponse(response, responseObject)
	} else {
		return handleNonProtoResponse(response)
	}
}

func handleProtoResponse(response *http.Response, responseObject proto.Message) error {
	fmt.Println("handleProtoResponse")
	if responseObject == nil {
		return models.NewError(models.Error_InvalidRequest, "responseObject cannot be nil")
	}

	fmt.Printf("response.Body: %+v\n", response.Body)
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return models.NewError(models.Error_InvalidResponse, fmt.Sprint("failed to read body: ", err.Error()))
	}

	fmt.Printf("responseObject (empty): %+v\n", responseObject)
	err = proto.Unmarshal(buf, responseObject)
	fmt.Printf("responseObject (filled): %+v\n", responseObject)
	if err != nil {
		return models.NewError(models.Error_InvalidProtobufMessage, fmt.Sprint("failed to unmarshal proto: ", err.Error()))
	}

	return nil
}

func handleNonProtoResponse(response *http.Response) error {
	if response.StatusCode == 404 {
		return EndpointNotFoundErr
	}

	if response.StatusCode > 299 {
		return models.NewError(models.Error_InvalidResponse, fmt.Sprintf(InvalidResponseMessage, response.StatusCode))
	}
	return nil
}
