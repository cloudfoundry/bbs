// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v4.25.6
// source: actions.proto

package models

// Prevent copylock errors when using ProtoAction directly
type Action struct {
	DownloadAction     *DownloadAction     `json:"download,omitempty"`
	UploadAction       *UploadAction       `json:"upload,omitempty"`
	RunAction          *RunAction          `json:"run,omitempty"`
	TimeoutAction      *TimeoutAction      `json:"timeout,omitempty"`
	EmitProgressAction *EmitProgressAction `json:"emit_progress,omitempty"`
	TryAction          *TryAction          `json:"try,omitempty"`
	ParallelAction     *ParallelAction     `json:"parallel,omitempty"`
	SerialAction       *SerialAction       `json:"serial,omitempty"`
	CodependentAction  *CodependentAction  `json:"codependent,omitempty"`
}

func (this *Action) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Action)
	if !ok {
		that2, ok := that.(Action)
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

	if this.DownloadAction == nil {
		if that1.DownloadAction != nil {
			return false
		}
	} else if !this.DownloadAction.Equal(*that1.DownloadAction) {
		return false
	}
	if this.UploadAction == nil {
		if that1.UploadAction != nil {
			return false
		}
	} else if !this.UploadAction.Equal(*that1.UploadAction) {
		return false
	}
	if this.RunAction == nil {
		if that1.RunAction != nil {
			return false
		}
	} else if !this.RunAction.Equal(*that1.RunAction) {
		return false
	}
	if this.TimeoutAction == nil {
		if that1.TimeoutAction != nil {
			return false
		}
	} else if !this.TimeoutAction.Equal(*that1.TimeoutAction) {
		return false
	}
	if this.EmitProgressAction == nil {
		if that1.EmitProgressAction != nil {
			return false
		}
	} else if !this.EmitProgressAction.Equal(*that1.EmitProgressAction) {
		return false
	}
	if this.TryAction == nil {
		if that1.TryAction != nil {
			return false
		}
	} else if !this.TryAction.Equal(*that1.TryAction) {
		return false
	}
	if this.ParallelAction == nil {
		if that1.ParallelAction != nil {
			return false
		}
	} else if !this.ParallelAction.Equal(*that1.ParallelAction) {
		return false
	}
	if this.SerialAction == nil {
		if that1.SerialAction != nil {
			return false
		}
	} else if !this.SerialAction.Equal(*that1.SerialAction) {
		return false
	}
	if this.CodependentAction == nil {
		if that1.CodependentAction != nil {
			return false
		}
	} else if !this.CodependentAction.Equal(*that1.CodependentAction) {
		return false
	}
	return true
}
func (m *Action) GetDownloadAction() *DownloadAction {
	if m != nil {
		return m.DownloadAction
	}
	return nil
}
func (m *Action) SetDownloadAction(value *DownloadAction) {
	if m != nil {
		m.DownloadAction = value
	}
}
func (m *Action) GetUploadAction() *UploadAction {
	if m != nil {
		return m.UploadAction
	}
	return nil
}
func (m *Action) SetUploadAction(value *UploadAction) {
	if m != nil {
		m.UploadAction = value
	}
}
func (m *Action) GetRunAction() *RunAction {
	if m != nil {
		return m.RunAction
	}
	return nil
}
func (m *Action) SetRunAction(value *RunAction) {
	if m != nil {
		m.RunAction = value
	}
}
func (m *Action) GetTimeoutAction() *TimeoutAction {
	if m != nil {
		return m.TimeoutAction
	}
	return nil
}
func (m *Action) SetTimeoutAction(value *TimeoutAction) {
	if m != nil {
		m.TimeoutAction = value
	}
}
func (m *Action) GetEmitProgressAction() *EmitProgressAction {
	if m != nil {
		return m.EmitProgressAction
	}
	return nil
}
func (m *Action) SetEmitProgressAction(value *EmitProgressAction) {
	if m != nil {
		m.EmitProgressAction = value
	}
}
func (m *Action) GetTryAction() *TryAction {
	if m != nil {
		return m.TryAction
	}
	return nil
}
func (m *Action) SetTryAction(value *TryAction) {
	if m != nil {
		m.TryAction = value
	}
}
func (m *Action) GetParallelAction() *ParallelAction {
	if m != nil {
		return m.ParallelAction
	}
	return nil
}
func (m *Action) SetParallelAction(value *ParallelAction) {
	if m != nil {
		m.ParallelAction = value
	}
}
func (m *Action) GetSerialAction() *SerialAction {
	if m != nil {
		return m.SerialAction
	}
	return nil
}
func (m *Action) SetSerialAction(value *SerialAction) {
	if m != nil {
		m.SerialAction = value
	}
}
func (m *Action) GetCodependentAction() *CodependentAction {
	if m != nil {
		return m.CodependentAction
	}
	return nil
}
func (m *Action) SetCodependentAction(value *CodependentAction) {
	if m != nil {
		m.CodependentAction = value
	}
}
func (x *Action) ToProto() *ProtoAction {
	if x == nil {
		return nil
	}

	proto := &ProtoAction{
		DownloadAction:     x.DownloadAction.ToProto(),
		UploadAction:       x.UploadAction.ToProto(),
		RunAction:          x.RunAction.ToProto(),
		TimeoutAction:      x.TimeoutAction.ToProto(),
		EmitProgressAction: x.EmitProgressAction.ToProto(),
		TryAction:          x.TryAction.ToProto(),
		ParallelAction:     x.ParallelAction.ToProto(),
		SerialAction:       x.SerialAction.ToProto(),
		CodependentAction:  x.CodependentAction.ToProto(),
	}
	return proto
}

func (x *ProtoAction) FromProto() *Action {
	if x == nil {
		return nil
	}

	copysafe := &Action{
		DownloadAction:     x.DownloadAction.FromProto(),
		UploadAction:       x.UploadAction.FromProto(),
		RunAction:          x.RunAction.FromProto(),
		TimeoutAction:      x.TimeoutAction.FromProto(),
		EmitProgressAction: x.EmitProgressAction.FromProto(),
		TryAction:          x.TryAction.FromProto(),
		ParallelAction:     x.ParallelAction.FromProto(),
		SerialAction:       x.SerialAction.FromProto(),
		CodependentAction:  x.CodependentAction.FromProto(),
	}
	return copysafe
}

func ActionToProtoSlice(values []*Action) []*ProtoAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ActionFromProtoSlice(values []*ProtoAction) []*Action {
	if values == nil {
		return nil
	}
	result := make([]*Action, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDownloadAction directly
type DownloadAction struct {
	Artifact          string `json:"artifact,omitempty"`
	From              string `json:"from"`
	To                string `json:"to"`
	CacheKey          string `json:"cache_key"`
	LogSource         string `json:"log_source,omitempty"`
	User              string `json:"user"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`
	ChecksumValue     string `json:"checksum_value,omitempty"`
}

func (this *DownloadAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DownloadAction)
	if !ok {
		that2, ok := that.(DownloadAction)
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

	if this.Artifact != that1.Artifact {
		return false
	}
	if this.From != that1.From {
		return false
	}
	if this.To != that1.To {
		return false
	}
	if this.CacheKey != that1.CacheKey {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.User != that1.User {
		return false
	}
	if this.ChecksumAlgorithm != that1.ChecksumAlgorithm {
		return false
	}
	if this.ChecksumValue != that1.ChecksumValue {
		return false
	}
	return true
}
func (m *DownloadAction) GetArtifact() string {
	if m != nil {
		return m.Artifact
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetArtifact(value string) {
	if m != nil {
		m.Artifact = value
	}
}
func (m *DownloadAction) GetFrom() string {
	if m != nil {
		return m.From
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetFrom(value string) {
	if m != nil {
		m.From = value
	}
}
func (m *DownloadAction) GetTo() string {
	if m != nil {
		return m.To
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetTo(value string) {
	if m != nil {
		m.To = value
	}
}
func (m *DownloadAction) GetCacheKey() string {
	if m != nil {
		return m.CacheKey
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetCacheKey(value string) {
	if m != nil {
		m.CacheKey = value
	}
}
func (m *DownloadAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *DownloadAction) GetUser() string {
	if m != nil {
		return m.User
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetUser(value string) {
	if m != nil {
		m.User = value
	}
}
func (m *DownloadAction) GetChecksumAlgorithm() string {
	if m != nil {
		return m.ChecksumAlgorithm
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetChecksumAlgorithm(value string) {
	if m != nil {
		m.ChecksumAlgorithm = value
	}
}
func (m *DownloadAction) GetChecksumValue() string {
	if m != nil {
		return m.ChecksumValue
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DownloadAction) SetChecksumValue(value string) {
	if m != nil {
		m.ChecksumValue = value
	}
}
func (x *DownloadAction) ToProto() *ProtoDownloadAction {
	if x == nil {
		return nil
	}

	proto := &ProtoDownloadAction{
		Artifact:          x.Artifact,
		From:              x.From,
		To:                x.To,
		CacheKey:          x.CacheKey,
		LogSource:         x.LogSource,
		User:              x.User,
		ChecksumAlgorithm: x.ChecksumAlgorithm,
		ChecksumValue:     x.ChecksumValue,
	}
	return proto
}

func (x *ProtoDownloadAction) FromProto() *DownloadAction {
	if x == nil {
		return nil
	}

	copysafe := &DownloadAction{
		Artifact:          x.Artifact,
		From:              x.From,
		To:                x.To,
		CacheKey:          x.CacheKey,
		LogSource:         x.LogSource,
		User:              x.User,
		ChecksumAlgorithm: x.ChecksumAlgorithm,
		ChecksumValue:     x.ChecksumValue,
	}
	return copysafe
}

func DownloadActionToProtoSlice(values []*DownloadAction) []*ProtoDownloadAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDownloadAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DownloadActionFromProtoSlice(values []*ProtoDownloadAction) []*DownloadAction {
	if values == nil {
		return nil
	}
	result := make([]*DownloadAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoUploadAction directly
type UploadAction struct {
	Artifact  string `json:"artifact,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	LogSource string `json:"log_source,omitempty"`
	User      string `json:"user"`
}

func (this *UploadAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*UploadAction)
	if !ok {
		that2, ok := that.(UploadAction)
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

	if this.Artifact != that1.Artifact {
		return false
	}
	if this.From != that1.From {
		return false
	}
	if this.To != that1.To {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.User != that1.User {
		return false
	}
	return true
}
func (m *UploadAction) GetArtifact() string {
	if m != nil {
		return m.Artifact
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UploadAction) SetArtifact(value string) {
	if m != nil {
		m.Artifact = value
	}
}
func (m *UploadAction) GetFrom() string {
	if m != nil {
		return m.From
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UploadAction) SetFrom(value string) {
	if m != nil {
		m.From = value
	}
}
func (m *UploadAction) GetTo() string {
	if m != nil {
		return m.To
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UploadAction) SetTo(value string) {
	if m != nil {
		m.To = value
	}
}
func (m *UploadAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UploadAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *UploadAction) GetUser() string {
	if m != nil {
		return m.User
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *UploadAction) SetUser(value string) {
	if m != nil {
		m.User = value
	}
}
func (x *UploadAction) ToProto() *ProtoUploadAction {
	if x == nil {
		return nil
	}

	proto := &ProtoUploadAction{
		Artifact:  x.Artifact,
		From:      x.From,
		To:        x.To,
		LogSource: x.LogSource,
		User:      x.User,
	}
	return proto
}

func (x *ProtoUploadAction) FromProto() *UploadAction {
	if x == nil {
		return nil
	}

	copysafe := &UploadAction{
		Artifact:  x.Artifact,
		From:      x.From,
		To:        x.To,
		LogSource: x.LogSource,
		User:      x.User,
	}
	return copysafe
}

func UploadActionToProtoSlice(values []*UploadAction) []*ProtoUploadAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoUploadAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func UploadActionFromProtoSlice(values []*ProtoUploadAction) []*UploadAction {
	if values == nil {
		return nil
	}
	result := make([]*UploadAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRunAction directly
type RunAction struct {
	Path              string                 `json:"path"`
	Args              []string               `json:"args,omitempty"`
	Dir               string                 `json:"dir,omitempty"`
	Env               []*EnvironmentVariable `json:"env,omitempty"`
	ResourceLimits    *ResourceLimits        `json:"resource_limits,omitempty"`
	User              string                 `json:"user"`
	LogSource         string                 `json:"log_source,omitempty"`
	SuppressLogOutput bool                   `json:"suppress_log_output"`
}

func (this *RunAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RunAction)
	if !ok {
		that2, ok := that.(RunAction)
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

	if this.Path != that1.Path {
		return false
	}
	if this.Args == nil {
		if that1.Args != nil {
			return false
		}
	} else if len(this.Args) != len(that1.Args) {
		return false
	}
	for i := range this.Args {
		if this.Args[i] != that1.Args[i] {
			return false
		}
	}
	if this.Dir != that1.Dir {
		return false
	}
	if this.Env == nil {
		if that1.Env != nil {
			return false
		}
	} else if len(this.Env) != len(that1.Env) {
		return false
	}
	for i := range this.Env {
		if !this.Env[i].Equal(that1.Env[i]) {
			return false
		}
	}
	if this.ResourceLimits == nil {
		if that1.ResourceLimits != nil {
			return false
		}
	} else if !this.ResourceLimits.Equal(*that1.ResourceLimits) {
		return false
	}
	if this.User != that1.User {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.SuppressLogOutput != that1.SuppressLogOutput {
		return false
	}
	return true
}
func (m *RunAction) GetPath() string {
	if m != nil {
		return m.Path
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RunAction) SetPath(value string) {
	if m != nil {
		m.Path = value
	}
}
func (m *RunAction) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}
func (m *RunAction) SetArgs(value []string) {
	if m != nil {
		m.Args = value
	}
}
func (m *RunAction) GetDir() string {
	if m != nil {
		return m.Dir
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RunAction) SetDir(value string) {
	if m != nil {
		m.Dir = value
	}
}
func (m *RunAction) GetEnv() []*EnvironmentVariable {
	if m != nil {
		return m.Env
	}
	return nil
}
func (m *RunAction) SetEnv(value []*EnvironmentVariable) {
	if m != nil {
		m.Env = value
	}
}
func (m *RunAction) GetResourceLimits() *ResourceLimits {
	if m != nil {
		return m.ResourceLimits
	}
	return nil
}
func (m *RunAction) SetResourceLimits(value *ResourceLimits) {
	if m != nil {
		m.ResourceLimits = value
	}
}
func (m *RunAction) GetUser() string {
	if m != nil {
		return m.User
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RunAction) SetUser(value string) {
	if m != nil {
		m.User = value
	}
}
func (m *RunAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RunAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *RunAction) GetSuppressLogOutput() bool {
	if m != nil {
		return m.SuppressLogOutput
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *RunAction) SetSuppressLogOutput(value bool) {
	if m != nil {
		m.SuppressLogOutput = value
	}
}
func (x *RunAction) ToProto() *ProtoRunAction {
	if x == nil {
		return nil
	}

	proto := &ProtoRunAction{
		Path:              x.Path,
		Args:              x.Args,
		Dir:               x.Dir,
		Env:               EnvironmentVariableToProtoSlice(x.Env),
		ResourceLimits:    x.ResourceLimits.ToProto(),
		User:              x.User,
		LogSource:         x.LogSource,
		SuppressLogOutput: x.SuppressLogOutput,
	}
	return proto
}

func (x *ProtoRunAction) FromProto() *RunAction {
	if x == nil {
		return nil
	}

	copysafe := &RunAction{
		Path:              x.Path,
		Args:              x.Args,
		Dir:               x.Dir,
		Env:               EnvironmentVariableFromProtoSlice(x.Env),
		ResourceLimits:    x.ResourceLimits.FromProto(),
		User:              x.User,
		LogSource:         x.LogSource,
		SuppressLogOutput: x.SuppressLogOutput,
	}
	return copysafe
}

func RunActionToProtoSlice(values []*RunAction) []*ProtoRunAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRunAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RunActionFromProtoSlice(values []*ProtoRunAction) []*RunAction {
	if values == nil {
		return nil
	}
	result := make([]*RunAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTimeoutAction directly
type TimeoutAction struct {
	Action *Action `json:"action,omitempty"`
	// Deprecated: marked deprecated in actions.proto
	DeprecatedTimeoutNs int64  `json:"timeout,omitempty"`
	LogSource           string `json:"log_source,omitempty"`
	TimeoutMs           int64  `json:"timeout_ms"`
}

func (this *TimeoutAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TimeoutAction)
	if !ok {
		that2, ok := that.(TimeoutAction)
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

	if this.Action == nil {
		if that1.Action != nil {
			return false
		}
	} else if !this.Action.Equal(*that1.Action) {
		return false
	}
	if this.DeprecatedTimeoutNs != that1.DeprecatedTimeoutNs {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	if this.TimeoutMs != that1.TimeoutMs {
		return false
	}
	return true
}
func (m *TimeoutAction) GetAction() *Action {
	if m != nil {
		return m.Action
	}
	return nil
}
func (m *TimeoutAction) SetAction(value *Action) {
	if m != nil {
		m.Action = value
	}
}

// Deprecated: marked deprecated in actions.proto
func (m *TimeoutAction) GetDeprecatedTimeoutNs() int64 {
	if m != nil {
		return m.DeprecatedTimeoutNs
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}

// Deprecated: marked deprecated in actions.proto
func (m *TimeoutAction) SetDeprecatedTimeoutNs(value int64) {
	if m != nil {
		m.DeprecatedTimeoutNs = value
	}
}
func (m *TimeoutAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TimeoutAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (m *TimeoutAction) GetTimeoutMs() int64 {
	if m != nil {
		return m.TimeoutMs
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}
func (m *TimeoutAction) SetTimeoutMs(value int64) {
	if m != nil {
		m.TimeoutMs = value
	}
}
func (x *TimeoutAction) ToProto() *ProtoTimeoutAction {
	if x == nil {
		return nil
	}

	proto := &ProtoTimeoutAction{
		Action:              x.Action.ToProto(),
		DeprecatedTimeoutNs: x.DeprecatedTimeoutNs,
		LogSource:           x.LogSource,
		TimeoutMs:           x.TimeoutMs,
	}
	return proto
}

func (x *ProtoTimeoutAction) FromProto() *TimeoutAction {
	if x == nil {
		return nil
	}

	copysafe := &TimeoutAction{
		Action:              x.Action.FromProto(),
		DeprecatedTimeoutNs: x.DeprecatedTimeoutNs,
		LogSource:           x.LogSource,
		TimeoutMs:           x.TimeoutMs,
	}
	return copysafe
}

func TimeoutActionToProtoSlice(values []*TimeoutAction) []*ProtoTimeoutAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTimeoutAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TimeoutActionFromProtoSlice(values []*ProtoTimeoutAction) []*TimeoutAction {
	if values == nil {
		return nil
	}
	result := make([]*TimeoutAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoEmitProgressAction directly
type EmitProgressAction struct {
	Action               *Action `json:"action,omitempty"`
	StartMessage         string  `json:"start_message"`
	SuccessMessage       string  `json:"success_message"`
	FailureMessagePrefix string  `json:"failure_message_prefix"`
	LogSource            string  `json:"log_source,omitempty"`
}

func (this *EmitProgressAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EmitProgressAction)
	if !ok {
		that2, ok := that.(EmitProgressAction)
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

	if this.Action == nil {
		if that1.Action != nil {
			return false
		}
	} else if !this.Action.Equal(*that1.Action) {
		return false
	}
	if this.StartMessage != that1.StartMessage {
		return false
	}
	if this.SuccessMessage != that1.SuccessMessage {
		return false
	}
	if this.FailureMessagePrefix != that1.FailureMessagePrefix {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	return true
}
func (m *EmitProgressAction) GetAction() *Action {
	if m != nil {
		return m.Action
	}
	return nil
}
func (m *EmitProgressAction) SetAction(value *Action) {
	if m != nil {
		m.Action = value
	}
}
func (m *EmitProgressAction) GetStartMessage() string {
	if m != nil {
		return m.StartMessage
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EmitProgressAction) SetStartMessage(value string) {
	if m != nil {
		m.StartMessage = value
	}
}
func (m *EmitProgressAction) GetSuccessMessage() string {
	if m != nil {
		return m.SuccessMessage
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EmitProgressAction) SetSuccessMessage(value string) {
	if m != nil {
		m.SuccessMessage = value
	}
}
func (m *EmitProgressAction) GetFailureMessagePrefix() string {
	if m != nil {
		return m.FailureMessagePrefix
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EmitProgressAction) SetFailureMessagePrefix(value string) {
	if m != nil {
		m.FailureMessagePrefix = value
	}
}
func (m *EmitProgressAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *EmitProgressAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (x *EmitProgressAction) ToProto() *ProtoEmitProgressAction {
	if x == nil {
		return nil
	}

	proto := &ProtoEmitProgressAction{
		Action:               x.Action.ToProto(),
		StartMessage:         x.StartMessage,
		SuccessMessage:       x.SuccessMessage,
		FailureMessagePrefix: x.FailureMessagePrefix,
		LogSource:            x.LogSource,
	}
	return proto
}

func (x *ProtoEmitProgressAction) FromProto() *EmitProgressAction {
	if x == nil {
		return nil
	}

	copysafe := &EmitProgressAction{
		Action:               x.Action.FromProto(),
		StartMessage:         x.StartMessage,
		SuccessMessage:       x.SuccessMessage,
		FailureMessagePrefix: x.FailureMessagePrefix,
		LogSource:            x.LogSource,
	}
	return copysafe
}

func EmitProgressActionToProtoSlice(values []*EmitProgressAction) []*ProtoEmitProgressAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoEmitProgressAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func EmitProgressActionFromProtoSlice(values []*ProtoEmitProgressAction) []*EmitProgressAction {
	if values == nil {
		return nil
	}
	result := make([]*EmitProgressAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTryAction directly
type TryAction struct {
	Action    *Action `json:"action,omitempty"`
	LogSource string  `json:"log_source,omitempty"`
}

func (this *TryAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TryAction)
	if !ok {
		that2, ok := that.(TryAction)
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

	if this.Action == nil {
		if that1.Action != nil {
			return false
		}
	} else if !this.Action.Equal(*that1.Action) {
		return false
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	return true
}
func (m *TryAction) GetAction() *Action {
	if m != nil {
		return m.Action
	}
	return nil
}
func (m *TryAction) SetAction(value *Action) {
	if m != nil {
		m.Action = value
	}
}
func (m *TryAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TryAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (x *TryAction) ToProto() *ProtoTryAction {
	if x == nil {
		return nil
	}

	proto := &ProtoTryAction{
		Action:    x.Action.ToProto(),
		LogSource: x.LogSource,
	}
	return proto
}

func (x *ProtoTryAction) FromProto() *TryAction {
	if x == nil {
		return nil
	}

	copysafe := &TryAction{
		Action:    x.Action.FromProto(),
		LogSource: x.LogSource,
	}
	return copysafe
}

func TryActionToProtoSlice(values []*TryAction) []*ProtoTryAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTryAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TryActionFromProtoSlice(values []*ProtoTryAction) []*TryAction {
	if values == nil {
		return nil
	}
	result := make([]*TryAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoParallelAction directly
type ParallelAction struct {
	Actions   []*Action `json:"actions,omitempty"`
	LogSource string    `json:"log_source,omitempty"`
}

func (this *ParallelAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ParallelAction)
	if !ok {
		that2, ok := that.(ParallelAction)
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

	if this.Actions == nil {
		if that1.Actions != nil {
			return false
		}
	} else if len(this.Actions) != len(that1.Actions) {
		return false
	}
	for i := range this.Actions {
		if !this.Actions[i].Equal(that1.Actions[i]) {
			return false
		}
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	return true
}
func (m *ParallelAction) GetActions() []*Action {
	if m != nil {
		return m.Actions
	}
	return nil
}
func (m *ParallelAction) SetActions(value []*Action) {
	if m != nil {
		m.Actions = value
	}
}
func (m *ParallelAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *ParallelAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (x *ParallelAction) ToProto() *ProtoParallelAction {
	if x == nil {
		return nil
	}

	proto := &ProtoParallelAction{
		Actions:   ActionToProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return proto
}

func (x *ProtoParallelAction) FromProto() *ParallelAction {
	if x == nil {
		return nil
	}

	copysafe := &ParallelAction{
		Actions:   ActionFromProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return copysafe
}

func ParallelActionToProtoSlice(values []*ParallelAction) []*ProtoParallelAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoParallelAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ParallelActionFromProtoSlice(values []*ProtoParallelAction) []*ParallelAction {
	if values == nil {
		return nil
	}
	result := make([]*ParallelAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoSerialAction directly
type SerialAction struct {
	Actions   []*Action `json:"actions,omitempty"`
	LogSource string    `json:"log_source,omitempty"`
}

func (this *SerialAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SerialAction)
	if !ok {
		that2, ok := that.(SerialAction)
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

	if this.Actions == nil {
		if that1.Actions != nil {
			return false
		}
	} else if len(this.Actions) != len(that1.Actions) {
		return false
	}
	for i := range this.Actions {
		if !this.Actions[i].Equal(that1.Actions[i]) {
			return false
		}
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	return true
}
func (m *SerialAction) GetActions() []*Action {
	if m != nil {
		return m.Actions
	}
	return nil
}
func (m *SerialAction) SetActions(value []*Action) {
	if m != nil {
		m.Actions = value
	}
}
func (m *SerialAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *SerialAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (x *SerialAction) ToProto() *ProtoSerialAction {
	if x == nil {
		return nil
	}

	proto := &ProtoSerialAction{
		Actions:   ActionToProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return proto
}

func (x *ProtoSerialAction) FromProto() *SerialAction {
	if x == nil {
		return nil
	}

	copysafe := &SerialAction{
		Actions:   ActionFromProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return copysafe
}

func SerialActionToProtoSlice(values []*SerialAction) []*ProtoSerialAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoSerialAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func SerialActionFromProtoSlice(values []*ProtoSerialAction) []*SerialAction {
	if values == nil {
		return nil
	}
	result := make([]*SerialAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCodependentAction directly
type CodependentAction struct {
	Actions   []*Action `json:"actions,omitempty"`
	LogSource string    `json:"log_source,omitempty"`
}

func (this *CodependentAction) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CodependentAction)
	if !ok {
		that2, ok := that.(CodependentAction)
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

	if this.Actions == nil {
		if that1.Actions != nil {
			return false
		}
	} else if len(this.Actions) != len(that1.Actions) {
		return false
	}
	for i := range this.Actions {
		if !this.Actions[i].Equal(that1.Actions[i]) {
			return false
		}
	}
	if this.LogSource != that1.LogSource {
		return false
	}
	return true
}
func (m *CodependentAction) GetActions() []*Action {
	if m != nil {
		return m.Actions
	}
	return nil
}
func (m *CodependentAction) SetActions(value []*Action) {
	if m != nil {
		m.Actions = value
	}
}
func (m *CodependentAction) GetLogSource() string {
	if m != nil {
		return m.LogSource
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CodependentAction) SetLogSource(value string) {
	if m != nil {
		m.LogSource = value
	}
}
func (x *CodependentAction) ToProto() *ProtoCodependentAction {
	if x == nil {
		return nil
	}

	proto := &ProtoCodependentAction{
		Actions:   ActionToProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return proto
}

func (x *ProtoCodependentAction) FromProto() *CodependentAction {
	if x == nil {
		return nil
	}

	copysafe := &CodependentAction{
		Actions:   ActionFromProtoSlice(x.Actions),
		LogSource: x.LogSource,
	}
	return copysafe
}

func CodependentActionToProtoSlice(values []*CodependentAction) []*ProtoCodependentAction {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCodependentAction, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CodependentActionFromProtoSlice(values []*ProtoCodependentAction) []*CodependentAction {
	if values == nil {
		return nil
	}
	result := make([]*CodependentAction, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoResourceLimits directly
type ResourceLimits struct {
	Nofile *uint64 `json:"nofile,omitempty"`
	// Deprecated: marked deprecated in actions.proto
	Nproc *uint64 `json:"nproc,omitempty"`
}

func (this *ResourceLimits) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ResourceLimits)
	if !ok {
		that2, ok := that.(ResourceLimits)
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

	if this.Nofile == nil {
		if that1.Nofile != nil {
			return false
		}
	} else if *this.Nofile != *that1.Nofile {
		return false
	}
	if this.Nproc == nil {
		if that1.Nproc != nil {
			return false
		}
	} else if *this.Nproc != *that1.Nproc {
		return false
	}
	return true
}
func (m *ResourceLimits) NofileExists() bool {
	return m != nil && m.Nofile != nil
}
func (m *ResourceLimits) GetNofile() *uint64 {
	if m != nil && m.Nofile != nil {
		return m.Nofile
	}
	return nil
}
func (m *ResourceLimits) SetNofile(value *uint64) {
	if m != nil {
		m.Nofile = value
	}
}

// Deprecated: marked deprecated in actions.proto
func (m *ResourceLimits) NprocExists() bool {
	return m != nil && m.Nproc != nil
}
func (m *ResourceLimits) GetNproc() *uint64 {
	if m != nil && m.Nproc != nil {
		return m.Nproc
	}
	return nil
}

// Deprecated: marked deprecated in actions.proto
func (m *ResourceLimits) SetNproc(value *uint64) {
	if m != nil {
		m.Nproc = value
	}
}
func (x *ResourceLimits) ToProto() *ProtoResourceLimits {
	if x == nil {
		return nil
	}

	proto := &ProtoResourceLimits{
		Nofile: x.Nofile,
		Nproc:  x.Nproc,
	}
	return proto
}

func (x *ProtoResourceLimits) FromProto() *ResourceLimits {
	if x == nil {
		return nil
	}

	copysafe := &ResourceLimits{
		Nofile: x.Nofile,
		Nproc:  x.Nproc,
	}
	return copysafe
}

func ResourceLimitsToProtoSlice(values []*ResourceLimits) []*ProtoResourceLimits {
	if values == nil {
		return nil
	}
	result := make([]*ProtoResourceLimits, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ResourceLimitsFromProtoSlice(values []*ProtoResourceLimits) []*ResourceLimits {
	if values == nil {
		return nil
	}
	result := make([]*ResourceLimits, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
