package models

// Auctioneer types moved here from code.cloudfoundry.org/auctioneer
// to break the bbs <-> auctioneer module cycle.

import (
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/lager/v3"
)

// AuctioneerClient is the interface bbs uses to communicate with the auctioneer.
// auctioneer implements this interface. Defined here to break the bbs <-> auctioneer cycle.
//
//go:generate counterfeiter -o fakes/fake_auctioneer_client.go . AuctioneerClient
type AuctioneerClient interface {
	RequestLRPAuctions(logger lager.Logger, traceID string, lrpStart []*LRPStartRequest) error
	RequestTaskAuctions(logger lager.Logger, traceID string, tasks []*TaskStartRequest) error
}

type TaskStartRequest struct {
	Task SchedulingTask
}

func NewTaskStartRequest(task SchedulingTask) TaskStartRequest {
	return TaskStartRequest{Task: task}
}

func NewTaskStartRequestFromModel(taskGuid, domain string, taskDef *TaskDefinition) TaskStartRequest {
	volumeMounts := []string{}
	for _, volumeMount := range taskDef.VolumeMounts {
		volumeMounts = append(volumeMounts, volumeMount.Driver)
	}
	return TaskStartRequest{
		Task: NewSchedulingTask(
			taskGuid,
			domain,
			NewResource(taskDef.MemoryMb, taskDef.DiskMb, taskDef.MaxPids),
			NewPlacementConstraint(taskDef.RootFs, taskDef.PlacementTags, volumeMounts),
		),
	}
}

// MarshalJSON flattens TaskStartRequest to the shape of its wrapped
// SchedulingTask. Before this type moved here from auctioneer, it
// anonymously embedded rep.Task, so its fields were promoted to the top
// level of the JSON object on the wire. Task became a named field in the
// move, which nested the payload under a "Task" key instead -- silently
// breaking auctions between an old auctioneer and a new BBS (or vice
// versa) during a rolling deploy, since the old side never finds its
// fields and drops every request with an empty-guid validation error.
// This preserves the pre-move wire format so mixed-version deploys stay
// compatible.
func (t TaskStartRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Task)
}

// UnmarshalJSON accepts the flat wire format described above.
func (t *TaskStartRequest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.Task)
}

func (t *TaskStartRequest) Validate() error {
	switch {
	case t.Task.TaskGuid == "":
		return errors.New("task guid is empty")
	case !t.Task.Resource.Valid():
		return errors.New("resources cannot be less than zero")
	case !t.Task.PlacementConstraint.Valid():
		return errors.New("placement constraint cannot be empty")
	default:
		return nil
	}
}

type LRPStartRequest struct {
	ProcessGuid string `json:"process_guid"`
	Domain      string `json:"domain"`
	Indices     []int  `json:"indices"`
	PlacementConstraint
	Resource
}

func NewLRPStartRequest(processGuid, domain string, indices []int, res Resource, pl PlacementConstraint) LRPStartRequest {
	return LRPStartRequest{
		ProcessGuid:         processGuid,
		Domain:              domain,
		Indices:             indices,
		Resource:            res,
		PlacementConstraint: pl,
	}
}

func NewLRPStartRequestFromModel(d *DesiredLRP, indices ...int) LRPStartRequest {
	volumeDrivers := []string{}
	for _, volumeMount := range d.VolumeMounts {
		volumeDrivers = append(volumeDrivers, volumeMount.Driver)
	}

	return NewLRPStartRequest(
		d.ProcessGuid,
		d.Domain,
		indices,
		NewResource(d.MemoryMb, d.DiskMb, d.MaxPids),
		NewPlacementConstraint(d.RootFs, d.PlacementTags, volumeDrivers),
	)
}

func NewLRPStartRequestFromSchedulingInfo(s *DesiredLRPSchedulingInfo, indices ...int) LRPStartRequest {
	return NewLRPStartRequest(
		s.ProcessGuid,
		s.Domain,
		indices,
		NewResource(s.MemoryMb, s.DiskMb, s.MaxPids),
		NewPlacementConstraint(s.RootFs, s.PlacementTags, s.VolumePlacement.DriverNames),
	)
}

func (lrpstart *LRPStartRequest) Validate() error {
	switch {
	case lrpstart.ProcessGuid == "":
		return errors.New("process guid is empty")
	case lrpstart.Domain == "":
		return errors.New("domain is empty")
	case len(lrpstart.Indices) == 0:
		return errors.New("indices must not be empty")
	case !lrpstart.Resource.Valid():
		return errors.New("resources cannot be less than 0")
	case !lrpstart.PlacementConstraint.Valid():
		return errors.New("placement constraint cannot be empty")
	default:
		return nil
	}
}
