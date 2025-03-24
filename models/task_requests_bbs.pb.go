// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.29.4
// source: task_requests.proto

package models

// Prevent copylock errors when using ProtoTaskLifecycleResponse directly
type TaskLifecycleResponse struct {
	Error *Error `json:"error,omitempty"`
}

func (this *TaskLifecycleResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskLifecycleResponse)
	if !ok {
		that2, ok := that.(TaskLifecycleResponse)
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	return true
}
func (m *TaskLifecycleResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *TaskLifecycleResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (x *TaskLifecycleResponse) ToProto() *ProtoTaskLifecycleResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskLifecycleResponse{
		Error: x.Error.ToProto(),
	}
	return proto
}

func (x *ProtoTaskLifecycleResponse) FromProto() *TaskLifecycleResponse {
	if x == nil {
		return nil
	}

	copysafe := &TaskLifecycleResponse{
		Error: x.Error.FromProto(),
	}
	return copysafe
}

func TaskLifecycleResponseToProtoSlice(values []*TaskLifecycleResponse) []*ProtoTaskLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskLifecycleResponseFromProtoSlice(values []*ProtoTaskLifecycleResponse) []*TaskLifecycleResponse {
	if values == nil {
		return nil
	}
	result := make([]*TaskLifecycleResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoDesireTaskRequest directly
type DesireTaskRequest struct {
	TaskDefinition *TaskDefinition `json:"task_definition"`
	TaskGuid       string          `json:"task_guid"`
	Domain         string          `json:"domain"`
}

func (this *DesireTaskRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DesireTaskRequest)
	if !ok {
		that2, ok := that.(DesireTaskRequest)
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
	return true
}
func (m *DesireTaskRequest) GetTaskDefinition() *TaskDefinition {
	if m != nil {
		return m.TaskDefinition
	}
	return nil
}
func (m *DesireTaskRequest) SetTaskDefinition(value *TaskDefinition) {
	if m != nil {
		m.TaskDefinition = value
	}
}
func (m *DesireTaskRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DesireTaskRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *DesireTaskRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *DesireTaskRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (x *DesireTaskRequest) ToProto() *ProtoDesireTaskRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoDesireTaskRequest{
		TaskDefinition: x.TaskDefinition.ToProto(),
		TaskGuid:       x.TaskGuid,
		Domain:         x.Domain,
	}
	return proto
}

func (x *ProtoDesireTaskRequest) FromProto() *DesireTaskRequest {
	if x == nil {
		return nil
	}

	copysafe := &DesireTaskRequest{
		TaskDefinition: x.TaskDefinition.FromProto(),
		TaskGuid:       x.TaskGuid,
		Domain:         x.Domain,
	}
	return copysafe
}

func DesireTaskRequestToProtoSlice(values []*DesireTaskRequest) []*ProtoDesireTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoDesireTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func DesireTaskRequestFromProtoSlice(values []*ProtoDesireTaskRequest) []*DesireTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*DesireTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoStartTaskRequest directly
type StartTaskRequest struct {
	TaskGuid string `json:"task_guid"`
	CellId   string `json:"cell_id"`
}

func (this *StartTaskRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*StartTaskRequest)
	if !ok {
		that2, ok := that.(StartTaskRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	return true
}
func (m *StartTaskRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *StartTaskRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *StartTaskRequest) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *StartTaskRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (x *StartTaskRequest) ToProto() *ProtoStartTaskRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoStartTaskRequest{
		TaskGuid: x.TaskGuid,
		CellId:   x.CellId,
	}
	return proto
}

func (x *ProtoStartTaskRequest) FromProto() *StartTaskRequest {
	if x == nil {
		return nil
	}

	copysafe := &StartTaskRequest{
		TaskGuid: x.TaskGuid,
		CellId:   x.CellId,
	}
	return copysafe
}

func StartTaskRequestToProtoSlice(values []*StartTaskRequest) []*ProtoStartTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoStartTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func StartTaskRequestFromProtoSlice(values []*ProtoStartTaskRequest) []*StartTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*StartTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoStartTaskResponse directly
type StartTaskResponse struct {
	Error       *Error `json:"error,omitempty"`
	ShouldStart bool   `json:"should_start"`
}

func (this *StartTaskResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*StartTaskResponse)
	if !ok {
		that2, ok := that.(StartTaskResponse)
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.ShouldStart != that1.ShouldStart {
		return false
	}
	return true
}
func (m *StartTaskResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *StartTaskResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *StartTaskResponse) GetShouldStart() bool {
	if m != nil {
		return m.ShouldStart
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *StartTaskResponse) SetShouldStart(value bool) {
	if m != nil {
		m.ShouldStart = value
	}
}
func (x *StartTaskResponse) ToProto() *ProtoStartTaskResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoStartTaskResponse{
		Error:       x.Error.ToProto(),
		ShouldStart: x.ShouldStart,
	}
	return proto
}

func (x *ProtoStartTaskResponse) FromProto() *StartTaskResponse {
	if x == nil {
		return nil
	}

	copysafe := &StartTaskResponse{
		Error:       x.Error.FromProto(),
		ShouldStart: x.ShouldStart,
	}
	return copysafe
}

func StartTaskResponseToProtoSlice(values []*StartTaskResponse) []*ProtoStartTaskResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoStartTaskResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func StartTaskResponseFromProtoSlice(values []*ProtoStartTaskResponse) []*StartTaskResponse {
	if values == nil {
		return nil
	}
	result := make([]*StartTaskResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Deprecated: marked deprecated in task_requests.proto
// Prevent copylock errors when using ProtoFailTaskRequest directly
type FailTaskRequest struct {
	TaskGuid      string `json:"task_guid"`
	FailureReason string `json:"failure_reason"`
}

func (this *FailTaskRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*FailTaskRequest)
	if !ok {
		that2, ok := that.(FailTaskRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.FailureReason != that1.FailureReason {
		return false
	}
	return true
}
func (m *FailTaskRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *FailTaskRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *FailTaskRequest) GetFailureReason() string {
	if m != nil {
		return m.FailureReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *FailTaskRequest) SetFailureReason(value string) {
	if m != nil {
		m.FailureReason = value
	}
}
func (x *FailTaskRequest) ToProto() *ProtoFailTaskRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoFailTaskRequest{
		TaskGuid:      x.TaskGuid,
		FailureReason: x.FailureReason,
	}
	return proto
}

func (x *ProtoFailTaskRequest) FromProto() *FailTaskRequest {
	if x == nil {
		return nil
	}

	copysafe := &FailTaskRequest{
		TaskGuid:      x.TaskGuid,
		FailureReason: x.FailureReason,
	}
	return copysafe
}

func FailTaskRequestToProtoSlice(values []*FailTaskRequest) []*ProtoFailTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoFailTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func FailTaskRequestFromProtoSlice(values []*ProtoFailTaskRequest) []*FailTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*FailTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoRejectTaskRequest directly
type RejectTaskRequest struct {
	TaskGuid        string `json:"task_guid"`
	RejectionReason string `json:"rejection_reason"`
}

func (this *RejectTaskRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RejectTaskRequest)
	if !ok {
		that2, ok := that.(RejectTaskRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.RejectionReason != that1.RejectionReason {
		return false
	}
	return true
}
func (m *RejectTaskRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RejectTaskRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *RejectTaskRequest) GetRejectionReason() string {
	if m != nil {
		return m.RejectionReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *RejectTaskRequest) SetRejectionReason(value string) {
	if m != nil {
		m.RejectionReason = value
	}
}
func (x *RejectTaskRequest) ToProto() *ProtoRejectTaskRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoRejectTaskRequest{
		TaskGuid:        x.TaskGuid,
		RejectionReason: x.RejectionReason,
	}
	return proto
}

func (x *ProtoRejectTaskRequest) FromProto() *RejectTaskRequest {
	if x == nil {
		return nil
	}

	copysafe := &RejectTaskRequest{
		TaskGuid:        x.TaskGuid,
		RejectionReason: x.RejectionReason,
	}
	return copysafe
}

func RejectTaskRequestToProtoSlice(values []*RejectTaskRequest) []*ProtoRejectTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoRejectTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func RejectTaskRequestFromProtoSlice(values []*ProtoRejectTaskRequest) []*RejectTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*RejectTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskGuidRequest directly
type TaskGuidRequest struct {
	TaskGuid string `json:"task_guid"`
}

func (this *TaskGuidRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskGuidRequest)
	if !ok {
		that2, ok := that.(TaskGuidRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	return true
}
func (m *TaskGuidRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskGuidRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (x *TaskGuidRequest) ToProto() *ProtoTaskGuidRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskGuidRequest{
		TaskGuid: x.TaskGuid,
	}
	return proto
}

func (x *ProtoTaskGuidRequest) FromProto() *TaskGuidRequest {
	if x == nil {
		return nil
	}

	copysafe := &TaskGuidRequest{
		TaskGuid: x.TaskGuid,
	}
	return copysafe
}

func TaskGuidRequestToProtoSlice(values []*TaskGuidRequest) []*ProtoTaskGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskGuidRequestFromProtoSlice(values []*ProtoTaskGuidRequest) []*TaskGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*TaskGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoCompleteTaskRequest directly
type CompleteTaskRequest struct {
	TaskGuid      string `json:"task_guid"`
	CellId        string `json:"cell_id"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Result        string `json:"result"`
}

func (this *CompleteTaskRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*CompleteTaskRequest)
	if !ok {
		that2, ok := that.(CompleteTaskRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	if this.Failed != that1.Failed {
		return false
	}
	if this.FailureReason != that1.FailureReason {
		return false
	}
	if this.Result != that1.Result {
		return false
	}
	return true
}
func (m *CompleteTaskRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CompleteTaskRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *CompleteTaskRequest) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CompleteTaskRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (m *CompleteTaskRequest) GetFailed() bool {
	if m != nil {
		return m.Failed
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *CompleteTaskRequest) SetFailed(value bool) {
	if m != nil {
		m.Failed = value
	}
}
func (m *CompleteTaskRequest) GetFailureReason() string {
	if m != nil {
		return m.FailureReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CompleteTaskRequest) SetFailureReason(value string) {
	if m != nil {
		m.FailureReason = value
	}
}
func (m *CompleteTaskRequest) GetResult() string {
	if m != nil {
		return m.Result
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *CompleteTaskRequest) SetResult(value string) {
	if m != nil {
		m.Result = value
	}
}
func (x *CompleteTaskRequest) ToProto() *ProtoCompleteTaskRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoCompleteTaskRequest{
		TaskGuid:      x.TaskGuid,
		CellId:        x.CellId,
		Failed:        x.Failed,
		FailureReason: x.FailureReason,
		Result:        x.Result,
	}
	return proto
}

func (x *ProtoCompleteTaskRequest) FromProto() *CompleteTaskRequest {
	if x == nil {
		return nil
	}

	copysafe := &CompleteTaskRequest{
		TaskGuid:      x.TaskGuid,
		CellId:        x.CellId,
		Failed:        x.Failed,
		FailureReason: x.FailureReason,
		Result:        x.Result,
	}
	return copysafe
}

func CompleteTaskRequestToProtoSlice(values []*CompleteTaskRequest) []*ProtoCompleteTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoCompleteTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func CompleteTaskRequestFromProtoSlice(values []*ProtoCompleteTaskRequest) []*CompleteTaskRequest {
	if values == nil {
		return nil
	}
	result := make([]*CompleteTaskRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskCallbackResponse directly
type TaskCallbackResponse struct {
	TaskGuid      string `json:"task_guid"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Result        string `json:"result"`
	Annotation    string `json:"annotation,omitempty"`
	CreatedAt     int64  `json:"created_at"`
}

func (this *TaskCallbackResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskCallbackResponse)
	if !ok {
		that2, ok := that.(TaskCallbackResponse)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	if this.Failed != that1.Failed {
		return false
	}
	if this.FailureReason != that1.FailureReason {
		return false
	}
	if this.Result != that1.Result {
		return false
	}
	if this.Annotation != that1.Annotation {
		return false
	}
	if this.CreatedAt != that1.CreatedAt {
		return false
	}
	return true
}
func (m *TaskCallbackResponse) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskCallbackResponse) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (m *TaskCallbackResponse) GetFailed() bool {
	if m != nil {
		return m.Failed
	}
	var defaultValue bool
	defaultValue = false
	return defaultValue
}
func (m *TaskCallbackResponse) SetFailed(value bool) {
	if m != nil {
		m.Failed = value
	}
}
func (m *TaskCallbackResponse) GetFailureReason() string {
	if m != nil {
		return m.FailureReason
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskCallbackResponse) SetFailureReason(value string) {
	if m != nil {
		m.FailureReason = value
	}
}
func (m *TaskCallbackResponse) GetResult() string {
	if m != nil {
		return m.Result
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskCallbackResponse) SetResult(value string) {
	if m != nil {
		m.Result = value
	}
}
func (m *TaskCallbackResponse) GetAnnotation() string {
	if m != nil {
		return m.Annotation
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskCallbackResponse) SetAnnotation(value string) {
	if m != nil {
		m.Annotation = value
	}
}
func (m *TaskCallbackResponse) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	var defaultValue int64
	defaultValue = 0
	return defaultValue
}
func (m *TaskCallbackResponse) SetCreatedAt(value int64) {
	if m != nil {
		m.CreatedAt = value
	}
}
func (x *TaskCallbackResponse) ToProto() *ProtoTaskCallbackResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskCallbackResponse{
		TaskGuid:      x.TaskGuid,
		Failed:        x.Failed,
		FailureReason: x.FailureReason,
		Result:        x.Result,
		Annotation:    x.Annotation,
		CreatedAt:     x.CreatedAt,
	}
	return proto
}

func (x *ProtoTaskCallbackResponse) FromProto() *TaskCallbackResponse {
	if x == nil {
		return nil
	}

	copysafe := &TaskCallbackResponse{
		TaskGuid:      x.TaskGuid,
		Failed:        x.Failed,
		FailureReason: x.FailureReason,
		Result:        x.Result,
		Annotation:    x.Annotation,
		CreatedAt:     x.CreatedAt,
	}
	return copysafe
}

func TaskCallbackResponseToProtoSlice(values []*TaskCallbackResponse) []*ProtoTaskCallbackResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskCallbackResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskCallbackResponseFromProtoSlice(values []*ProtoTaskCallbackResponse) []*TaskCallbackResponse {
	if values == nil {
		return nil
	}
	result := make([]*TaskCallbackResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTasksRequest directly
type TasksRequest struct {
	Domain string `json:"domain"`
	CellId string `json:"cell_id"`
}

func (this *TasksRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TasksRequest)
	if !ok {
		that2, ok := that.(TasksRequest)
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

	if this.Domain != that1.Domain {
		return false
	}
	if this.CellId != that1.CellId {
		return false
	}
	return true
}
func (m *TasksRequest) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TasksRequest) SetDomain(value string) {
	if m != nil {
		m.Domain = value
	}
}
func (m *TasksRequest) GetCellId() string {
	if m != nil {
		return m.CellId
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TasksRequest) SetCellId(value string) {
	if m != nil {
		m.CellId = value
	}
}
func (x *TasksRequest) ToProto() *ProtoTasksRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoTasksRequest{
		Domain: x.Domain,
		CellId: x.CellId,
	}
	return proto
}

func (x *ProtoTasksRequest) FromProto() *TasksRequest {
	if x == nil {
		return nil
	}

	copysafe := &TasksRequest{
		Domain: x.Domain,
		CellId: x.CellId,
	}
	return copysafe
}

func TasksRequestToProtoSlice(values []*TasksRequest) []*ProtoTasksRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTasksRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TasksRequestFromProtoSlice(values []*ProtoTasksRequest) []*TasksRequest {
	if values == nil {
		return nil
	}
	result := make([]*TasksRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTasksResponse directly
type TasksResponse struct {
	Error *Error  `json:"error,omitempty"`
	Tasks []*Task `json:"tasks,omitempty"`
}

func (this *TasksResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TasksResponse)
	if !ok {
		that2, ok := that.(TasksResponse)
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.Tasks == nil {
		if that1.Tasks != nil {
			return false
		}
	} else if len(this.Tasks) != len(that1.Tasks) {
		return false
	}
	for i := range this.Tasks {
		if !this.Tasks[i].Equal(that1.Tasks[i]) {
			return false
		}
	}
	return true
}
func (m *TasksResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *TasksResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *TasksResponse) GetTasks() []*Task {
	if m != nil {
		return m.Tasks
	}
	return nil
}
func (m *TasksResponse) SetTasks(value []*Task) {
	if m != nil {
		m.Tasks = value
	}
}
func (x *TasksResponse) ToProto() *ProtoTasksResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoTasksResponse{
		Error: x.Error.ToProto(),
		Tasks: TaskToProtoSlice(x.Tasks),
	}
	return proto
}

func (x *ProtoTasksResponse) FromProto() *TasksResponse {
	if x == nil {
		return nil
	}

	copysafe := &TasksResponse{
		Error: x.Error.FromProto(),
		Tasks: TaskFromProtoSlice(x.Tasks),
	}
	return copysafe
}

func TasksResponseToProtoSlice(values []*TasksResponse) []*ProtoTasksResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTasksResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TasksResponseFromProtoSlice(values []*ProtoTasksResponse) []*TasksResponse {
	if values == nil {
		return nil
	}
	result := make([]*TasksResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskByGuidRequest directly
type TaskByGuidRequest struct {
	TaskGuid string `json:"task_guid"`
}

func (this *TaskByGuidRequest) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskByGuidRequest)
	if !ok {
		that2, ok := that.(TaskByGuidRequest)
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

	if this.TaskGuid != that1.TaskGuid {
		return false
	}
	return true
}
func (m *TaskByGuidRequest) GetTaskGuid() string {
	if m != nil {
		return m.TaskGuid
	}
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *TaskByGuidRequest) SetTaskGuid(value string) {
	if m != nil {
		m.TaskGuid = value
	}
}
func (x *TaskByGuidRequest) ToProto() *ProtoTaskByGuidRequest {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskByGuidRequest{
		TaskGuid: x.TaskGuid,
	}
	return proto
}

func (x *ProtoTaskByGuidRequest) FromProto() *TaskByGuidRequest {
	if x == nil {
		return nil
	}

	copysafe := &TaskByGuidRequest{
		TaskGuid: x.TaskGuid,
	}
	return copysafe
}

func TaskByGuidRequestToProtoSlice(values []*TaskByGuidRequest) []*ProtoTaskByGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskByGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskByGuidRequestFromProtoSlice(values []*ProtoTaskByGuidRequest) []*TaskByGuidRequest {
	if values == nil {
		return nil
	}
	result := make([]*TaskByGuidRequest, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}

// Prevent copylock errors when using ProtoTaskResponse directly
type TaskResponse struct {
	Error *Error `json:"error,omitempty"`
	Task  *Task  `json:"task,omitempty"`
}

func (this *TaskResponse) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TaskResponse)
	if !ok {
		that2, ok := that.(TaskResponse)
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

	if this.Error == nil {
		if that1.Error != nil {
			return false
		}
	} else if !this.Error.Equal(*that1.Error) {
		return false
	}
	if this.Task == nil {
		if that1.Task != nil {
			return false
		}
	} else if !this.Task.Equal(*that1.Task) {
		return false
	}
	return true
}
func (m *TaskResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}
func (m *TaskResponse) SetError(value *Error) {
	if m != nil {
		m.Error = value
	}
}
func (m *TaskResponse) GetTask() *Task {
	if m != nil {
		return m.Task
	}
	return nil
}
func (m *TaskResponse) SetTask(value *Task) {
	if m != nil {
		m.Task = value
	}
}
func (x *TaskResponse) ToProto() *ProtoTaskResponse {
	if x == nil {
		return nil
	}

	proto := &ProtoTaskResponse{
		Error: x.Error.ToProto(),
		Task:  x.Task.ToProto(),
	}
	return proto
}

func (x *ProtoTaskResponse) FromProto() *TaskResponse {
	if x == nil {
		return nil
	}

	copysafe := &TaskResponse{
		Error: x.Error.FromProto(),
		Task:  x.Task.FromProto(),
	}
	return copysafe
}

func TaskResponseToProtoSlice(values []*TaskResponse) []*ProtoTaskResponse {
	if values == nil {
		return nil
	}
	result := make([]*ProtoTaskResponse, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func TaskResponseFromProtoSlice(values []*ProtoTaskResponse) []*TaskResponse {
	if values == nil {
		return nil
	}
	result := make([]*TaskResponse, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
