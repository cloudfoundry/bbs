package executor

import (
	"errors"
	"time"

	"code.cloudfoundry.org/bbs/models"
)

const ContainerOwnerProperty = "executor:owner"

type State string

const (
	StateInvalid      State = ""
	StateReserved     State = "reserved"
	StateInitializing State = "initializing"
	StateCreated      State = "created"
	StateRunning      State = "running"
	StateCompleted    State = "completed"
)

const (
	HealthcheckTag      = "executor-healthcheck"
	HealthcheckTagValue = "executor-healthcheck"
)

type ProxyPortMapping struct {
	AppPort   uint16 `json:"app_port"`
	ProxyPort uint16 `json:"proxy_port"`
}

type Container struct {
	Guid string `json:"guid"`
	Resource
	RunInfo
	Tags                                  Tags
	State                                 State              `json:"state"`
	AllocatedAt                           int64              `json:"allocated_at"`
	ExternalIP                            string             `json:"external_ip"`
	InternalIP                            string             `json:"internal_ip"`
	RunResult                             ContainerRunResult `json:"run_result"`
	MemoryLimit                           uint64             `json:"memory_limit"`
	DiskLimit                             uint64             `json:"disk_limit"`
	AdvertisePreferenceForInstanceAddress bool               `json:"advertise_preference_for_instance_address"`
}

func NewContainerFromResource(guid string, resource *Resource, tags Tags) Container {
	return Container{
		Guid:     guid,
		Resource: *resource,
		Tags:     tags,
	}
}

func (c *Container) ValidateTransitionTo(newState State) bool {
	if newState == StateCompleted {
		return true
	}
	switch c.State {
	case StateReserved:
		return newState == StateInitializing
	case StateInitializing:
		return newState == StateCreated
	case StateCreated:
		return newState == StateRunning
	default:
		return false
	}
}

func (c *Container) TransistionToInitialize(req *RunRequest) error {
	if !c.ValidateTransitionTo(StateInitializing) {
		return ErrInvalidTransition
	}
	c.State = StateInitializing
	c.RunInfo = req.RunInfo
	c.Tags.Add(req.Tags)
	return nil
}

func (c *Container) TransitionToCreate() error {
	if !c.ValidateTransitionTo(StateCreated) {
		return ErrInvalidTransition
	}

	c.State = StateCreated
	return nil
}

func (c *Container) TransitionToComplete(failed bool, failureReason string, retryable bool) {
	c.RunResult.Failed = failed
	c.RunResult.FailureReason = failureReason
	c.RunResult.Retryable = retryable
	c.State = StateCompleted
}

func (newContainer Container) Copy() Container {
	newContainer.Tags = newContainer.Tags.Copy()
	return newContainer
}

func (c *Container) IsCreated() bool {
	return c.State != StateReserved && c.State != StateInitializing && c.State != StateCompleted
}

func (c *Container) HasTags(tags Tags) bool {
	if c.Tags == nil {
		return tags == nil
	}

	if tags == nil {
		return false
	}

	for key, val := range tags {
		v, ok := c.Tags[key]
		if !ok || val != v {
			return false
		}
	}

	return true
}

func NewReservedContainerFromAllocationRequest(req *AllocationRequest, allocatedAt int64) Container {
	c := NewContainerFromResource(req.Guid, &req.Resource, req.Tags)
	c.State = StateReserved
	c.AllocatedAt = allocatedAt
	return c
}

type Resource struct {
	MemoryMB int `json:"memory_mb"`
	DiskMB   int `json:"disk_mb"`
	MaxPids  int `json:"max_pids"`
}

func NewResource(memoryMB, diskMB, maxPids int) Resource {
	return Resource{
		MemoryMB: memoryMB,
		DiskMB:   diskMB,
		MaxPids:  maxPids,
	}
}

type CachedDependency struct {
	Name              string `json:"name"`
	From              string `json:"from"`
	To                string `json:"to"`
	CacheKey          string `json:"cache_key"`
	LogSource         string `json:"log_source"`
	ChecksumValue     string `json:"checksum_value"`
	ChecksumAlgorithm string `json:"checksum_algorithm"`
}

type CertificateProperties struct {
	OrganizationalUnit []string `json:"organizational_unit"`
}

type Sidecar struct {
	Action   *models.Action `json:"run"`
	DiskMB   int32          `json:"disk_mb"`
	MemoryMB int32          `json:"memory_mb"`
}

type RunInfo struct {
	RootFSPath                    string                      `json:"rootfs"`
	CPUWeight                     uint                        `json:"cpu_weight"`
	Ports                         []PortMapping               `json:"ports"`
	LogConfig                     LogConfig                   `json:"log_config"`
	MetricsConfig                 MetricsConfig               `json:"metrics_config"`
	StartTimeoutMs                uint                        `json:"start_timeout_ms"`
	Privileged                    bool                        `json:"privileged"`
	CachedDependencies            []CachedDependency          `json:"cached_dependencies"`
	Setup                         *models.Action              `json:"setup"`
	Action                        *models.Action              `json:"run"`
	Monitor                       *models.Action              `json:"monitor"`
	CheckDefinition               *models.CheckDefinition     `json:"check_definition"`
	EgressRules                   []*models.SecurityGroupRule `json:"egress_rules,omitempty"`
	Env                           []EnvironmentVariable       `json:"env,omitempty"`
	TrustedSystemCertificatesPath string                      `json:"trusted_system_certificates_path,omitempty"`
	VolumeMounts                  []VolumeMount               `json:"volume_mounts"`
	Network                       *Network                    `json:"network,omitempty"`
	CertificateProperties         CertificateProperties       `json:"certificate_properties"`
	ImageUsername                 string                      `json:"image_username"`
	ImagePassword                 string                      `json:"image_password"`
	EnableContainerProxy          bool                        `json:"enable_container_proxy"`
	Sidecars                      []Sidecar                   `json:"sidecars"`
}

type BindMountMode uint8

const (
	BindMountModeRO BindMountMode = 0
	BindMountModeRW BindMountMode = 1
)

type VolumeMount struct {
	Driver        string                 `json:"driver"`
	VolumeId      string                 `json:"volume_id"`
	Config        map[string]interface{} `json:"config"`
	ContainerPath string                 `json:"container_path"`
	Mode          BindMountMode          `json:"mode"`
}

type Network struct {
	Properties map[string]string `json:"properties,omitempty"`
}

type InnerContainer Container

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ContainerMetrics struct {
	MemoryUsageInBytes                  uint64        `json:"memory_usage_in_bytes"`
	DiskUsageInBytes                    uint64        `json:"disk_usage_in_bytes"`
	MemoryLimitInBytes                  uint64        `json:"memory_limit_in_bytes"`
	DiskLimitInBytes                    uint64        `json:"disk_limit_in_bytes"`
	TimeSpentInCPU                      time.Duration `json:"time_spent_in_cpu"`
	AbsoluteCPUEntitlementInNanoseconds uint64        `json:"absolute_cpu_entitlement_in_ns"`
	ContainerAgeInNanoseconds           uint64        `json:"container_age_in_ns"`
}

type MetricsConfig struct {
	Guid  string            `json:"guid"`
	Index int               `json:"index"`
	Tags  map[string]string `json:"tags"`
}

type Metrics struct {
	MetricsConfig
	ContainerMetrics
}

type LogConfig struct {
	Guid       string            `json:"guid"`
	Index      int               `json:"index"`
	SourceName string            `json:"source_name"`
	Tags       map[string]string `json:"tags"`
}

type PortMapping struct {
	ContainerPort         uint16 `json:"container_port"`
	HostPort              uint16 `json:"host_port,omitempty"`
	ContainerTLSProxyPort uint16 `json:"container_tls_proxy_port,omitempty"`
	HostTLSProxyPort      uint16 `json:"host_tls_proxy_port,omitempty"`
}

type ContainerRunResult struct {
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Retryable     bool

	Stopped bool `json:"stopped"`
}

type ExecutorResources struct {
	MemoryMB   int `json:"memory_mb"`
	DiskMB     int `json:"disk_mb"`
	Containers int `json:"containers"`
}

func NewExecutorResources(memoryMB, diskMB, containers int) ExecutorResources {
	return ExecutorResources{
		MemoryMB:   memoryMB,
		DiskMB:     diskMB,
		Containers: containers,
	}
}

func (e ExecutorResources) Copy() ExecutorResources {
	return e
}

func (r *ExecutorResources) canSubtract(res *Resource) bool {
	return r.MemoryMB >= res.MemoryMB && r.DiskMB >= res.DiskMB && r.Containers > 0
}

func (r *ExecutorResources) Subtract(res *Resource) bool {
	if !r.canSubtract(res) {
		return false
	}
	r.MemoryMB -= res.MemoryMB
	r.DiskMB -= res.DiskMB
	r.Containers -= 1
	return true
}

func (r *ExecutorResources) Add(res *Resource) {
	r.MemoryMB += res.MemoryMB
	r.DiskMB += res.DiskMB
	r.Containers += 1
}

type Tags map[string]string

func (t Tags) Copy() Tags {
	if t == nil {
		return nil
	}
	newTags := make(Tags, len(t))
	newTags.Add(t)
	return newTags
}

func (t Tags) Add(other Tags) {
	for key := range other {
		t[key] = other[key]
	}
}

type Event interface {
	EventType() EventType
}

type EventType string

var ErrUnknownEventType = errors.New("unknown event type")

const (
	EventTypeInvalid EventType = ""

	EventTypeContainerComplete EventType = "container_complete"
	EventTypeContainerRunning  EventType = "container_running"
	EventTypeContainerReserved EventType = "container_reserved"
)

type LifecycleEvent interface {
	Container() Container
	lifecycleEvent()
}

type ContainerCompleteEvent struct {
	RawContainer Container `json:"container"`
}

func NewContainerCompleteEvent(container Container) ContainerCompleteEvent {
	return ContainerCompleteEvent{
		RawContainer: container,
	}
}

func (ContainerCompleteEvent) EventType() EventType   { return EventTypeContainerComplete }
func (e ContainerCompleteEvent) Container() Container { return e.RawContainer }
func (ContainerCompleteEvent) lifecycleEvent()        {}

type ContainerRunningEvent struct {
	RawContainer Container `json:"container"`
}

func NewContainerRunningEvent(container Container) ContainerRunningEvent {
	return ContainerRunningEvent{
		RawContainer: container,
	}
}

func (ContainerRunningEvent) EventType() EventType   { return EventTypeContainerRunning }
func (e ContainerRunningEvent) Container() Container { return e.RawContainer }
func (ContainerRunningEvent) lifecycleEvent()        {}

type ContainerReservedEvent struct {
	RawContainer Container `json:"container"`
}

func NewContainerReservedEvent(container Container) ContainerReservedEvent {
	return ContainerReservedEvent{
		RawContainer: container,
	}
}

func (ContainerReservedEvent) EventType() EventType   { return EventTypeContainerReserved }
func (e ContainerReservedEvent) Container() Container { return e.RawContainer }
func (ContainerReservedEvent) lifecycleEvent()        {}
