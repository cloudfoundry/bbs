package rep

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/executor/containermetrics"
)

var ErrorIncompatibleRootfs = errors.New("rootfs not found")

type CellState struct {
	RepURL                  string `json:"rep_url"`
	CellID                  string `json:"cell_id"`
	CellIndex               int    `json:"cell_index"`
	RootFSProviders         RootFSProviders
	AvailableResources      Resources
	TotalResources          Resources
	LRPs                    []LRP
	Tasks                   []Task
	StartingContainerCount  int
	Zone                    string
	Evacuating              bool
	VolumeDrivers           []string
	PlacementTags           []string
	OptionalPlacementTags   []string
	ProxyMemoryAllocationMB int
}

func NewCellState(
	cellID string,
	cellIndex int,
	repURL string,
	root RootFSProviders,
	avail Resources,
	total Resources,
	lrps []LRP,
	tasks []Task,
	zone string,
	startingContainerCount int,
	isEvac bool,
	volumeDrivers []string,
	placementTags []string,
	optionalPlacementTags []string,
	proxyMemoryAllocation int,
) CellState {
	return CellState{
		CellID:                  cellID,
		CellIndex:               cellIndex,
		RepURL:                  repURL,
		RootFSProviders:         root,
		AvailableResources:      avail,
		TotalResources:          total,
		LRPs:                    lrps,
		Tasks:                   tasks,
		Zone:                    zone,
		StartingContainerCount:  startingContainerCount,
		Evacuating:              isEvac,
		VolumeDrivers:           volumeDrivers,
		PlacementTags:           placementTags,
		OptionalPlacementTags:   optionalPlacementTags,
		ProxyMemoryAllocationMB: proxyMemoryAllocation,
	}
}

func (c *CellState) AddLRP(lrp *LRP) {
	c.AvailableResources.Subtract(&lrp.Resource)
	c.StartingContainerCount += 1
	c.LRPs = append(c.LRPs, *lrp)
}

func (c *CellState) AddTask(task *Task) {
	c.AvailableResources.Subtract(&task.Resource)
	c.StartingContainerCount += 1
	c.Tasks = append(c.Tasks, *task)
}

func (c *CellState) ResourceMatch(res *Resource) error {
	problems := map[string]struct{}{}

	if c.AvailableResources.DiskMB < res.DiskMB {
		problems["disk"] = struct{}{}
	}
	if c.AvailableResources.MemoryMB < res.MemoryMB {
		problems["memory"] = struct{}{}
	}
	if c.AvailableResources.Containers < 1 {
		problems["containers"] = struct{}{}
	}
	if len(problems) == 0 {
		return nil
	}

	return InsufficientResourcesError{Problems: problems}
}

type InsufficientResourcesError struct {
	Problems map[string]struct{}
}

func (i InsufficientResourcesError) Error() string {
	if len(i.Problems) == 0 {
		return "insufficient resources"
	}

	keys := []string{}
	for key, _ := range i.Problems {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return fmt.Sprintf("insufficient resources: %s", strings.Join(keys, ", "))
}

func (c CellState) ComputeScore(res *Resource, startingContainerWeight float64) float64 {
	remainingResources := c.AvailableResources.Copy()
	remainingResources.Subtract(res)
	startingContainerScore := float64(c.StartingContainerCount) * startingContainerWeight
	return remainingResources.ComputeScore(&c.TotalResources) + startingContainerScore
}

func (c *CellState) MatchRootFS(rootfs string) bool {
	rootFSURL, err := url.Parse(rootfs)
	if err != nil {
		return false
	}

	return c.RootFSProviders.Match(*rootFSURL)
}

func (c *CellState) MatchVolumeDrivers(volumeDrivers []string) bool {
	for _, requestedDriver := range volumeDrivers {
		found := false

		for _, actualDriver := range c.VolumeDrivers {
			if requestedDriver == actualDriver {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (c *CellState) MatchPlacementTags(desiredPlacementTags []string) bool {
	desiredTags := toSet(desiredPlacementTags)
	optionalTags := toSet(c.OptionalPlacementTags)
	requiredTags := toSet(c.PlacementTags)
	allTags := requiredTags.union(optionalTags)

	return requiredTags.isSubset(desiredTags) && desiredTags.isSubset(allTags)
}

type placementTagSet map[string]struct{}

func (set placementTagSet) union(other placementTagSet) placementTagSet {
	tags := placementTagSet{}
	for k := range set {
		tags[k] = struct{}{}
	}
	for k := range other {
		tags[k] = struct{}{}
	}
	return tags
}

func (set placementTagSet) isSubset(other placementTagSet) bool {
	for k := range set {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

func toSet(slice []string) placementTagSet {
	tags := placementTagSet{}
	for _, k := range slice {
		tags[k] = struct{}{}
	}
	return tags
}

type Resources struct {
	MemoryMB   int32
	DiskMB     int32
	Containers int
}

func NewResources(memoryMb, diskMb int32, containerCount int) Resources {
	return Resources{memoryMb, diskMb, containerCount}
}

func (r *Resources) Copy() Resources {
	return *r
}

func (r *Resources) Subtract(res *Resource) {
	r.MemoryMB -= res.MemoryMB
	r.DiskMB -= res.DiskMB
	r.Containers -= 1
}

func (r *Resources) ComputeScore(total *Resources) float64 {
	fractionUsedMemory := 1.0 - float64(r.MemoryMB)/float64(total.MemoryMB)
	fractionUsedDisk := 1.0 - float64(r.DiskMB)/float64(total.DiskMB)
	fractionUsedContainers := 1.0 - float64(r.Containers)/float64(total.Containers)
	return (fractionUsedMemory + fractionUsedDisk + fractionUsedContainers) / 3.0
}

type Resource struct {
	MemoryMB int32
	DiskMB   int32
	MaxPids  int32
}

func NewResource(memoryMb, diskMb int32, maxPids int32) Resource {
	return Resource{MemoryMB: memoryMb, DiskMB: diskMb, MaxPids: maxPids}
}

func (r *Resource) Valid() bool {
	return r.DiskMB >= 0 && r.MemoryMB >= 0
}

func (r *Resource) Copy() Resource {
	return NewResource(r.MemoryMB, r.DiskMB, r.MaxPids)
}

type PlacementConstraint struct {
	PlacementTags []string
	VolumeDrivers []string
	RootFs        string
}

func NewPlacementConstraint(rootFs string, placementTags, volumeDrivers []string) PlacementConstraint {
	return PlacementConstraint{PlacementTags: placementTags, VolumeDrivers: volumeDrivers, RootFs: rootFs}
}

func (p *PlacementConstraint) Valid() bool {
	return p.RootFs != ""
}

type LRP struct {
	InstanceGUID string `json:"instance_guid"`
	models.ActualLRPKey
	PlacementConstraint
	Resource
	State string `json:"state"`
}

func NewLRP(instanceGUID string, key models.ActualLRPKey, res Resource, pc PlacementConstraint) LRP {
	return LRP{instanceGUID, key, pc, res, ""}
}

func (lrp *LRP) Identifier() string {
	return fmt.Sprintf("%s.%d", lrp.ProcessGuid, lrp.Index)
}

func (lrp *LRP) Copy() LRP {
	return NewLRP(lrp.InstanceGUID, lrp.ActualLRPKey, lrp.Resource, lrp.PlacementConstraint)
}

type Task struct {
	TaskGuid string
	Domain   string
	PlacementConstraint
	Resource
	State  models.Task_State `json:"state"`
	Failed bool              `json:"failed"`
}

func NewTask(guid string, domain string, res Resource, pc PlacementConstraint) Task {
	return Task{guid, domain, pc, res, models.Task_Invalid, false}
}

func (task *Task) Identifier() string {
	return task.TaskGuid
}

func (task Task) Copy() Task {
	return task
}

type Work struct {
	LRPs   []LRP
	Tasks  []Task
	CellID string `json:"cell_id,omitempty"`
}

// StackPathMap maps aliases to rootFS paths on the system.
type StackPathMap map[string]string

// ErrPreloadedRootFSNotFound is returned when the given hostname of the
// rootFS could not be resolved if the scheme is the PreloadedRootFSScheme
// or the PreloadedOCIRootFSScheme. This isn't the error for when the actual
// path on the system could not be found.
var ErrPreloadedRootFSNotFound = errors.New("preloaded rootfs path not found")

// PathForRootFS resolves the hostname portion of the RootFS URL to the actual
// path to the preloaded rootFS on the system according to the StackPathMap
func (m StackPathMap) PathForRootFS(rootFS string) (string, error) {
	if rootFS == "" {
		return rootFS, nil
	}

	url, err := url.Parse(rootFS)
	if err != nil {
		return "", err
	}

	if url.Scheme == models.PreloadedRootFSScheme {
		path, ok := m[url.Opaque]
		if !ok {
			return "", ErrPreloadedRootFSNotFound
		}
		return path, nil
	} else if url.Scheme == models.PreloadedOCIRootFSScheme {
		path, ok := m[url.Opaque]
		if !ok {
			return "", ErrPreloadedRootFSNotFound
		}

		return fmt.Sprintf("%s:%s?%s", url.Scheme, path, url.RawQuery), nil
	}

	return rootFS, nil
}

//go:generate counterfeiter -o auctioncellrep/auctioncellrepfakes/fake_container_metrics_provider.go . ContainerMetricsProvider
type ContainerMetricsProvider interface {
	Metrics() map[string]*containermetrics.CachedContainerMetrics
}

type ContainerMetricsCollection struct {
	CellID string       `json:"cell_id"`
	LRPs   []LRPMetric  `json:"lrps"`
	Tasks  []TaskMetric `json:"tasks"`
}

type LRPMetric struct {
	InstanceGUID string `json:"instance_guid"`
	ProcessGUID  string `json:"process_guid"`
	Index        int32  `json:"index"`
	containermetrics.CachedContainerMetrics
}

type TaskMetric struct {
	TaskGUID string `json:"task_guid"`
	containermetrics.CachedContainerMetrics
}
