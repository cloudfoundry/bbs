package executor

import (
	"io"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o fakes/fake_client.go . Client

type Client interface {
	Ping(logger lager.Logger) error
	AllocateContainers(logger lager.Logger, requests []AllocationRequest) []AllocationFailure
	GetContainer(logger lager.Logger, guid string) (Container, error)
	RunContainer(lager.Logger, *RunRequest) error
	StopContainer(logger lager.Logger, guid string) error
	DeleteContainer(logger lager.Logger, guid string) error
	ListContainers(lager.Logger) ([]Container, error)
	GetBulkMetrics(lager.Logger) (map[string]Metrics, error)
	RemainingResources(lager.Logger) (ExecutorResources, error)
	TotalResources(lager.Logger) (ExecutorResources, error)
	GetFiles(logger lager.Logger, guid string, path string) (io.ReadCloser, error)
	VolumeDrivers(logger lager.Logger) ([]string, error)
	SubscribeToEvents(lager.Logger) (EventSource, error)
	Healthy(lager.Logger) bool
	SetHealthy(lager.Logger, bool)
	Cleanup(lager.Logger)
}

//go:generate counterfeiter -o fakes/fake_event_source.go . EventSource

type EventSource interface {
	Next() (Event, error)
	Close() error
}

type AllocationRequest struct {
	Guid string
	Resource
	Tags
}

func NewAllocationRequest(guid string, resource *Resource, tags Tags) AllocationRequest {
	return AllocationRequest{
		Guid:     guid,
		Resource: *resource,
		Tags:     tags,
	}
}

func (a *AllocationRequest) Validate() error {
	if a.Guid == "" {
		return ErrGuidNotSpecified
	}
	return nil
}

type AllocationFailure struct {
	AllocationRequest
	ErrorMsg string
}

func (fail *AllocationFailure) Error() string {
	return fail.ErrorMsg
}

func NewAllocationFailure(req *AllocationRequest, msg string) AllocationFailure {
	return AllocationFailure{
		AllocationRequest: *req,
		ErrorMsg:          msg,
	}
}

type RunRequest struct {
	Guid string
	RunInfo
	Tags
}

func NewRunRequest(guid string, runInfo *RunInfo, tags Tags) RunRequest {
	return RunRequest{
		Guid:    guid,
		RunInfo: *runInfo,
		Tags:    tags,
	}
}
