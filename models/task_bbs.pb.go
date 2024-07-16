// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.28.0
// source: task.proto

package models

import (
	strconv "strconv"
)

// Prevent copylock errors when using ProtoTaskDefinition directly
type TaskDefinition struct {
	RootFs                string                 `json:"rootfs"`
	EnvironmentVariables  []*EnvironmentVariable `json:"env,omitempty"`
	Action                *Action                `json:"action,omitempty"`
	DiskMb                int32                  `json:"disk_mb"`
	MemoryMb              int32                  `json:"memory_mb"`
	CpuWeight             uint32                 `json:"cpu_weight"`
	Privileged            bool                   `json:"privileged"`
	LogSource             string                 `json:"log_source"`
	LogGuid               string                 `json:"log_guid"`
	MetricsGuid           string                 `json:"metrics_guid"`
	ResultFile            string                 `json:"result_file"`
	CompletionCallbackUrl string                 `json:"completion_callback_url,omitempty"`
	Annotation            string                 `json:"annotation,omitempty"`
	EgressRules           []*SecurityGroupRule   `json:"egress_rules,omitempty"`
	CachedDependencies    []*CachedDependency    `json:"cached_dependencies,omitempty"`
	// Deprecated: marked deprecated in task.proto
	LegacyDownloadUser            string                     `json:"legacy_download_user,omitempty"`
	TrustedSystemCertificatesPath string                     `json:"trusted_system_certificates_path,omitempty"`
	VolumeMounts                  []*VolumeMount             `json:"volume_mounts,omitempty"`
	Network                       *Network                   `json:"network,omitempty"`
	PlacementTags                 []string                   `json:"placement_tags,omitempty"`
	MaxPids                       int32                      `json:"max_pids"`
	CertificateProperties         *CertificateProperties     `json:"certificate_properties,omitempty"`
	ImageUsername                 string                     `json:"image_username"`
	ImagePassword                 string                     `json:"image_password"`
	ImageLayers                   []*ImageLayer              `json:"image_layers,omitempty"`
	LogRateLimit                  *LogRateLimit              `json:"log_rate_limit,omitempty"`
	MetricTags                    map[string]*MetricTagValue `json:"metric_tags,omitempty"`
}

func (this *TaskDefinition) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskDefinition)
	if !ok {
		that2, ok := that.(TaskDefinition)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if this.RootFs != that1.RootFs {
		return false
	}
	if this.EnvironmentVariables == nil {
		if that1.EnvironmentVariables != nil {
			return false
		}
	} else if len(this.EnvironmentVariables) != len(that1.EnvironmentVariables) {
		return false
	}
	for i := range this.EnvironmentVariables {
		if !this.EnvironmentVariables[i].Equal(that1.EnvironmentVariables[i]) {
			return false
		}
	}
	if this.Action == nil {
		if that1.Action != nil {
			return false
		}
	} else if !this.Action.Equal(*that1.Action) {
		return false
	}
	if this.DiskMb != that1.DiskMb {
		return false
	}
	if this.MemoryMb != that1.MemoryMb {
		return false
	}
	if this.CpuWeight != that1.CpuWeight {
		return false
	}
	if this.Privileged != that1.Privileged {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.LogGuid != that1.LogGuid {
		return false
	}
	if this.MetricsGuid != that1.MetricsGuid {
		return false
	}
	if this.ResultFile != that1.ResultFile {
		return false
	}
	if this.CompletionCallbackUrl != that1.CompletionCallbackUrl {
		return false
	}
	if this.Annotation != that1.Annotation {
		return false
	}
	if this.EgressRules == nil {
		if that1.EgressRules != nil {
			return false
		}
	} else if len(this.EgressRules) != len(that1.EgressRules) {
		return false
	}
	for i := range this.EgressRules {
		if !this.EgressRules[i].Equal(that1.EgressRules[i]) {
			return false
		}
	}
	if this.CachedDependencies == nil {
		if that1.CachedDependencies != nil {
			return false
		}
	} else if len(this.CachedDependencies) != len(that1.CachedDependencies) {
		return false
	}
	for i := range this.CachedDependencies {
		if !this.CachedDependencies[i].Equal(that1.CachedDependencies[i]) {
			return false
		}
	}
	if this.LegacyDownloadUser != that1.LegacyDownloadUser {
		return false
	}
	if this.TrustedSystemCertificatesPath != that1.TrustedSystemCertificatesPath {
		return false
	}
	if this.VolumeMounts == nil {
		if that1.VolumeMounts != nil {
			return false
		}
	} else if len(this.VolumeMounts) != len(that1.VolumeMounts) {
		return false
	}
	for i := range this.VolumeMounts {
		if !this.VolumeMounts[i].Equal(that1.VolumeMounts[i]) {
			return false
		}
	}
	if this.Network == nil {
		if that1.Network != nil {
			return false
		}
	} else if !this.Network.Equal(*that1.Network) {
		return false
	}
	if this.PlacementTags == nil {
		if that1.PlacementTags != nil {
			return false
		}
	} else if len(this.PlacementTags) != len(that1.PlacementTags) {
		return false
	}
	for i := range this.PlacementTags {
		if this.PlacementTags[i] != that1.PlacementTags[i] {
			return false
		}
	}
	if this.MaxPids != that1.MaxPids {
		return false
	}
	if this.CertificateProperties == nil {
		if that1.CertificateProperties != nil {
			return false
		}
	} else if !this.CertificateProperties.Equal(*that1.CertificateProperties) {
		return false
	}
	if this.ImageUsername != that1.ImageUsername {
		return false
	}
	if this.ImagePassword != that1.ImagePassword {
		return false
	}
	if this.ImageLayers == nil {
		if that1.ImageLayers != nil {
			return false
		}
	} else if len(this.ImageLayers) != len(that1.ImageLayers) {
		return false
	}
	for i := range this.ImageLayers {
		if !this.ImageLayers[i].Equal(that1.ImageLayers[i]) {
			return false
		}
	}
	if this.LogRateLimit == nil {
		if that1.LogRateLimit != nil {
			return false
		}
	} else if !this.LogRateLimit.Equal(*that1.LogRateLimit) {
		return false
	}
	if this.MetricTags == nil {
		if that1.MetricTags != nil {
			return false
		}
	} else if len(this.MetricTags) != len(that1.MetricTags) {
		return false
	}
	for i := range this.MetricTags {
		if !this.MetricTags[i].Equal(that1.MetricTags[i]) {
			return false
		}
	}
	return true
}
func (m *TaskDefinition) GetRootFs() string {
	if m != nil {
		return m.RootFs
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetRootFs(value string) {
	if m != nil {
		m.RootFs = value
	}
}
func (m *TaskDefinition) GetEnvironmentVariables() []*EnvironmentVariable {
	if m != nil {
		return m.EnvironmentVariables
	}
	return nil
}
func (m *TaskDefinition) SetEnvironmentVariables(value []*EnvironmentVariable) {
	if m != nil {
		m.EnvironmentVariables = value
	}
}
func (m *TaskDefinition) GetAction() *Action {
	if m != nil {
		return m.Action
	}
	return nil
}
func (m *TaskDefinition) SetAction(value *Action) {
	if m != nil {
		m.Action = value
	}
}
func (m *TaskDefinition) GetDiskMb() int32 {
	if m != nil {
		return m.DiskMb
	}
	var defaultValue int32
	defaultValue = 0
	return defaultValue
}
func (m *TaskDefinition) SetDiskMb(value int32) {
	if m != nil {
		m.DiskMb = value
	}
}
func (m *TaskDefinition) GetMemoryMb() int32 {
	if m != nil {
		return m.MemoryMb
	}
	var defaultValue int32
	defaultValue = 0
	return defaultValue
}
func (m *TaskDefinition) SetMemoryMb(value int32) {
	if m != nil {
		m.MemoryMb = value
	}
}
func (m *TaskDefinition) GetCpuWeight() uint32 {
	if m != nil {
		return m.CpuWeight
	}
	var defaultValue uint32
	defaultValue = 0
	return defaultValue
}
func (m *TaskDefinition) SetCpuWeight(value uint32) {
	if m != nil {
		m.CpuWeight = value
	}
}
func (m *TaskDefinition) GetPrivileged() bool {
	if m != nil {
		return m.Privileged
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *TaskDefinition) SetPrivileged(value bool) {
	if m != nil {
		m.Privileged = value
	}
}
func (m *TaskDefinition) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *TaskDefinition) GetLogGuid() string {
	if m != nil {
		return m.LogGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetLogGuid(value string) {
	if m != nil {
		m.LogGuid = value
	}
}
func (m *TaskDefinition) GetMetricsGuid() string {
	if m != nil {
		return m.MetricsGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetMetricsGuid(value string) {
	if m != nil {
		m.MetricsGuid = value
	}
}
func (m *TaskDefinition) GetResultFile() string {
	if m != nil {
		return m.ResultFile
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetResultFile(value string) {
	if m != nil {
		m.ResultFile = value
	}
}
func (m *TaskDefinition) GetCompletionCallbackUrl() string {
	if m != nil {
		return m.CompletionCallbackUrl
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetCompletionCallbackUrl(value string) {
	if m != nil {
		m.CompletionCallbackUrl = value
	}
}
func (m *TaskDefinition) GetAnnotation() string {
	if m != nil {
		return m.Annotation
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetAnnotation(value string) {
	if m != nil {
		m.Annotation = value
	}
}
func (m *TaskDefinition) GetEgressRules() []*SecurityGroupRule {
	if m != nil {
		return m.EgressRules
	}
	return nil
}
func (m *TaskDefinition) SetEgressRules(value []*SecurityGroupRule) {
	if m != nil {
		m.EgressRules = value
	}
}
func (m *TaskDefinition) GetCachedDependencies() []*CachedDependency {
	if m != nil {
		return m.CachedDependencies
	}
	return nil
}
func (m *TaskDefinition) SetCachedDependencies(value []*CachedDependency) {
	if m != nil {
		m.CachedDependencies = value
	}
}

// Deprecated: marked deprecated in task.proto
func (m *TaskDefinition) GetLegacyDownloadUser() string {
	if m != nil {
		return m.LegacyDownloadUser
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}

// Deprecated: marked deprecated in task.proto
func (m *TaskDefinition) SetLegacyDownloadUser(value string) {
	if m != nil {
		m.LegacyDownloadUser = value
	}
}
func (m *TaskDefinition) GetTrustedSystemCertificatesPath() string {
	if m != nil {
		return m.TrustedSystemCertificatesPath
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetTrustedSystemCertificatesPath(value string) {
	if m != nil {
		m.TrustedSystemCertificatesPath = value
	}
}
func (m *TaskDefinition) GetVolumeMounts() []*VolumeMount {
	if m != nil {
		return m.VolumeMounts
	}
	return nil
}
func (m *TaskDefinition) SetVolumeMounts(value []*VolumeMount) {
	if m != nil {
		m.VolumeMounts = value
	}
}
func (m *TaskDefinition) GetNetwork() *Network {
	if m != nil {
		return m.Network
	}
	return nil
}
func (m *TaskDefinition) SetNetwork(value *Network) {
	if m != nil {
		m.Network = value
	}
}
func (m *TaskDefinition) GetPlacementTags() []string {
	if m != nil {
		return m.PlacementTags
	}
	return nil
}
func (m *TaskDefinition) SetPlacementTags(value []string) {
	if m != nil {
		m.PlacementTags = value
	}
}
func (m *TaskDefinition) GetMaxPids() int32 {
	if m != nil {
		return m.MaxPids
	}
	var defaultValue int32
	defaultValue = 0
	return defaultValue
}
func (m *TaskDefinition) SetMaxPids(value int32) {
	if m != nil {
		m.MaxPids = value
	}
}
func (m *TaskDefinition) GetCertificateProperties() *CertificateProperties {
	if m != nil {
		return m.CertificateProperties
	}
	return nil
}
func (m *TaskDefinition) SetCertificateProperties(value *CertificateProperties) {
	if m != nil {
		m.CertificateProperties = value
	}
}
func (m *TaskDefinition) GetImageUsername() string {
	if m != nil {
		return m.ImageUsername
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetImageUsername(value string) {
	if m != nil {
		m.ImageUsername = value
	}
}
func (m *TaskDefinition) GetImagePassword() string {
	if m != nil {
		return m.ImagePassword
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskDefinition) SetImagePassword(value string) {
	if m != nil {
		m.ImagePassword = value
	}
}
func (m *TaskDefinition) GetImageLayers() []*ImageLayer {
	if m != nil {
		return m.ImageLayers
	}
	return nil
}
func (m *TaskDefinition) SetImageLayers(value []*ImageLayer) {
	if m != nil {
		m.ImageLayers = value
	}
}
func (m *TaskDefinition) GetLogRateLimit() *LogRateLimit {
	if m != nil {
		return m.LogRateLimit
	}
	return nil
}
func (m *TaskDefinition) SetLogRateLimit(value *LogRateLimit) {
	if m != nil {
		m.LogRateLimit = value
	}
}
func (m *TaskDefinition) GetMetricTags() map[string]*MetricTagValue {
	if m != nil {
		return m.MetricTags
	}
	return nil
}
func (m *TaskDefinition) SetMetricTags(value map[string]*MetricTagValue) {
	if m != nil {
		m.MetricTags = value
	}
}
func (x *TaskDefinition) ToProto() *ProtoTaskDefinition {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskDefinition{
		RootFs:                        x.RootFs,
		EnvironmentVariables:          EnvironmentVariableToProtoSlice(x.EnvironmentVariables),
		Action:                        x.Action.ToProto(),
		DiskMb:                        x.DiskMb,
		MemoryMb:                      x.MemoryMb,
		CpuWeight:                     x.CpuWeight,
		Privileged:                    x.Privileged,
		LogSource:                     x.LogSource,
		LogGuid:                       x.LogGuid,
		MetricsGuid:                   x.MetricsGuid,
		ResultFile:                    x.ResultFile,
		CompletionCallbackUrl:         x.CompletionCallbackUrl,
		Annotation:                    x.Annotation,
		EgressRules:                   SecurityGroupRuleToProtoSlice(x.EgressRules),
		CachedDependencies:            CachedDependencyToProtoSlice(x.CachedDependencies),
		LegacyDownloadUser:            x.LegacyDownloadUser,
		TrustedSystemCertificatesPath: x.TrustedSystemCertificatesPath,
		VolumeMounts:                  VolumeMountToProtoSlice(x.VolumeMounts),
		Network:                       x.Network.ToProto(),
		PlacementTags:                 x.PlacementTags,
		MaxPids:                       x.MaxPids,
		CertificateProperties:         x.CertificateProperties.ToProto(),
		ImageUsername:                 x.ImageUsername,
		ImagePassword:                 x.ImagePassword,
		ImageLayers:                   ImageLayerToProtoSlice(x.ImageLayers),
		LogRateLimit:                  x.LogRateLimit.ToProto(),
		MetricTags:                    TaskDefinitionMetricTagsToProtoMap(x.MetricTags),
	}
	return proto
}

func (x *ProtoTaskDefinition) FromProto() *TaskDefinition {
	if x == nil {
		return nil
	}

	copysafe := &TaskDefinition{
		RootFs:                        x.RootFs,
		EnvironmentVariables:          EnvironmentVariableFromProtoSlice(x.EnvironmentVariables),
		Action:                        x.Action.FromProto(),
		DiskMb:                        x.DiskMb,
		MemoryMb:                      x.MemoryMb,
		CpuWeight:                     x.CpuWeight,
		Privileged:                    x.Privileged,
		LogSource:                     x.LogSource,
		LogGuid:                       x.LogGuid,
		MetricsGuid:                   x.MetricsGuid,
		ResultFile:                    x.ResultFile,
		CompletionCallbackUrl:         x.CompletionCallbackUrl,
		Annotation:                    x.Annotation,
		EgressRules:                   SecurityGroupRuleFromProtoSlice(x.EgressRules),
		CachedDependencies:            CachedDependencyFromProtoSlice(x.CachedDependencies),
		LegacyDownloadUser:            x.LegacyDownloadUser,
		TrustedSystemCertificatesPath: x.TrustedSystemCertificatesPath,
		VolumeMounts:                  VolumeMountFromProtoSlice(x.VolumeMounts),
		Network:                       x.Network.FromProto(),
		PlacementTags:                 x.PlacementTags,
		MaxPids:                       x.MaxPids,
		CertificateProperties:         x.CertificateProperties.FromProto(),
		ImageUsername:                 x.ImageUsername,
		ImagePassword:                 x.ImagePassword,
		ImageLayers:                   ImageLayerFromProtoSlice(x.ImageLayers),
		LogRateLimit:                  x.LogRateLimit.FromProto(),
		MetricTags:                    TaskDefinitionMetricTagsFromProtoMap(x.MetricTags),
	}
	return copysafe
}

func TaskDefinitionToProtoSlice(values []*TaskDefinition) []*ProtoTaskDefinition {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskDefinition, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskDefinitionMetricTagsToProtoMap(values map[string]*MetricTagValue) map[string]*ProtoMetricTagValue {
	if values == nil {
		return nil
	}
	result := make(map[string]*ProtoMetricTagValue, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskDefinitionFromProtoSlice(values []*ProtoTaskDefinition) []*TaskDefinition {
	if values == nil {
		return nil
	}
	result := make([]*TaskDefinition, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

func TaskDefinitionMetricTagsFromProtoMap(values map[string]*ProtoMetricTagValue) map[string]*MetricTagValue {
	if values == nil {
		return nil
	}
	result := make(map[string]*MetricTagValue, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

type Task_State int32

const (
	Task_Invalid   Task_State = 0
	Task_Pending   Task_State = 1
	Task_Running   Task_State = 2
	Task_Completed Task_State = 3
	Task_Resolving Task_State = 4
)

// Enum value maps for Task_State
var (
	Task_State_name = map[int32]string{
		0: "Invalid",
		1: "Pending",
		2: "Running",
		3: "Completed",
		4: "Resolving",
	}
	Task_State_value = map[string]int32{
		"Invalid":   0,
		"Pending":   1,
		"Running":   2,
		"Completed": 3,
		"Resolving": 4,
	}
)

func (m Task_State) String() string {
	s, ok := Task_State_name[int32(m)]
	if ok {
		return s
	}
	return strconv.Itoa(int(m))
}

// Prevent copylock errors when using ProtoTask directly
type Task struct {
	TaskDefinition   *TaskDefinition `json:"task_definition"`
	TaskGuid         string          `json:"task_guid"`
	Domain           string          `json:"domain"`
	CreatedAt        int64           `json:"created_at"`
	UpdatedAt        int64           `json:"updated_at"`
	FirstCompletedAt int64           `json:"first_completed_at"`
	State            Task_State      `json:"state"`
	CellId           string          `json:"cell_id"`
	Result           string          `json:"result"`
	Failed           bool            `json:"failed"`
	FailureReason    string          `json:"failure_reason"`
	RejectionCount   int32           `json:"rejection_count"`
	RejectionReason  string          `json:"rejection_reason"`
}

func (this *Task) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Task)
	if !ok {
		that2, ok := that.(Task)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if this.TaskDefinition == nil {
		if that1.TaskDefinition != nil {
			return false
		}
	} else if !this.TaskDefinition.Equal(*that1.TaskDefinition) {
		return false
	}
	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.Domain != that1.Domain {
		return false
	}
	if this.CreatedAt != that1.CreatedAt {
		return false
	}
	if this.UpdatedAt != that1.UpdatedAt {
		return false
	}
	if this.FirstCompletedAt != that1.FirstCompletedAt {
		return false
	}
	if this.State != that1.State {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	if this.Result != that1.Result {
		return false
	}
	if this.Failed != that1.Failed {
		return false
	}
	if this.FailureReason != that1.FailureReason {
		return false
	}
	if this.RejectionCount != that1.RejectionCount {
		return false
	}
	if this.RejectionReason != that1.RejectionReason {
		return false
	}
	return true
}
func (m *Task) GetTaskDefinition() *TaskDefinition {
	if m != nil {
		return m.TaskDefinition
	}
	return nil
}
func (m *Task) SetTaskDefinition(value *TaskDefinition) {
	if m != nil {
		m.TaskDefinition = value
	}
}
func (m *Task) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *Task) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *Task) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}
func (m *Task) SetCreatedAt(value int64) {
	if m != nil {
		m.CreatedAt = value
	}
}
func (m *Task) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}
func (m *Task) SetUpdatedAt(value int64) {
	if m != nil {
		m.UpdatedAt = value
	}
}
func (m *Task) GetFirstCompletedAt() int64 {
	if m != nil {
		return m.FirstCompletedAt
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}
func (m *Task) SetFirstCompletedAt(value int64) {
	if m != nil {
		m.FirstCompletedAt = value
	}
}
func (m *Task) GetState() Task_State {
	if m != nil {
		return m.State
	}
	var defaultValue Task_State
	defaultValue = 0
	return defaultValue
}
func (m *Task) SetState(value Task_State) {
	if m != nil {
		m.State = value
	}
}
func (m *Task) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (m *Task) GetResult() string {
	if m != nil {
		return m.Result
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetResult(value string) {
	if m != nil {
		m.Result = value
	}
}
func (m *Task) GetFailed() bool {
	if m != nil {
		return m.Failed
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *Task) SetFailed(value bool) {
	if m != nil {
		m.Failed = value
	}
}
func (m *Task) GetFailureReason() string {
	if m != nil {
		return m.FailureReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetFailureReason(value string) {
	if m != nil {
		m.FailureReason = value
	}
}
func (m *Task) GetRejectionCount() int32 {
	if m != nil {
		return m.RejectionCount
	}
	var defaultValue int32
	defaultValue = 0
	return defaultValue
}
func (m *Task) SetRejectionCount(value int32) {
	if m != nil {
		m.RejectionCount = value
	}
}
func (m *Task) GetRejectionReason() string {
	if m != nil {
		return m.RejectionReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Task) SetRejectionReason(value string) {
	if m != nil {
		m.RejectionReason = value
	}
}
func (x *Task) ToProto() *ProtoTask {
	if x == nil {
		return nil
	}

	proto := &ProtoTask{
		TaskDefinition:   x.TaskDefinition.ToProto(),
		TaskGuid:         x.TaskGuid,
		Domain:           x.Domain,
		CreatedAt:        x.CreatedAt,
		UpdatedAt:        x.UpdatedAt,
		FirstCompletedAt: x.FirstCompletedAt,
		State:            ProtoTask_State(x.State),
		CellId:           x.CellId,
		Result:           x.Result,
		Failed:           x.Failed,
		FailureReason:    x.FailureReason,
		RejectionCount:   x.RejectionCount,
		RejectionReason:  x.RejectionReason,
	}
	return proto
}

func (x *ProtoTask) FromProto() *Task {
	if x == nil {
		return nil
	}

	copysafe := &Task{
		TaskDefinition:   x.TaskDefinition.FromProto(),
		TaskGuid:         x.TaskGuid,
		Domain:           x.Domain,
		CreatedAt:        x.CreatedAt,
		UpdatedAt:        x.UpdatedAt,
		FirstCompletedAt: x.FirstCompletedAt,
		State:            Task_State(x.State),
		CellId:           x.CellId,
		Result:           x.Result,
		Failed:           x.Failed,
		FailureReason:    x.FailureReason,
		RejectionCount:   x.RejectionCount,
		RejectionReason:  x.RejectionReason,
	}
	return copysafe
}

func TaskToProtoSlice(values []*Task) []*ProtoTask {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTask, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskFromProtoSlice(values []*ProtoTask) []*Task {
	if values == nil {
		return nil
	}
	result := make([]*Task, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
