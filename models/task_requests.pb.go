// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.0
// source: task_requests.proto

package models

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ProtoTaskLifecycleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ProtoTaskLifecycleResponse) Reset() {
	*x = ProtoTaskLifecycleResponse{}
	mi := &file_task_requests_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTaskLifecycleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTaskLifecycleResponse) ProtoMessage() {}

func (x *ProtoTaskLifecycleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTaskLifecycleResponse.ProtoReflect.Descriptor instead.
func (*ProtoTaskLifecycleResponse) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoTaskLifecycleResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

type ProtoDesireTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskDefinition *ProtoTaskDefinition `protobuf:"bytes,1,opt,name=task_definition,proto3" json:"task_definition,omitempty"`
	TaskGuid       string               `protobuf:"bytes,2,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	Domain         string               `protobuf:"bytes,3,opt,name=domain,proto3" json:"domain,omitempty"`
}

func (x *ProtoDesireTaskRequest) Reset() {
	*x = ProtoDesireTaskRequest{}
	mi := &file_task_requests_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesireTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesireTaskRequest) ProtoMessage() {}

func (x *ProtoDesireTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesireTaskRequest.ProtoReflect.Descriptor instead.
func (*ProtoDesireTaskRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoDesireTaskRequest) GetTaskDefinition() *ProtoTaskDefinition {
	if x != nil {
		return x.TaskDefinition
	}
	return nil
}

func (x *ProtoDesireTaskRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoDesireTaskRequest) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type ProtoStartTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	CellId   string `protobuf:"bytes,2,opt,name=cell_id,proto3" json:"cell_id,omitempty"`
}

func (x *ProtoStartTaskRequest) Reset() {
	*x = ProtoStartTaskRequest{}
	mi := &file_task_requests_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoStartTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoStartTaskRequest) ProtoMessage() {}

func (x *ProtoStartTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoStartTaskRequest.ProtoReflect.Descriptor instead.
func (*ProtoStartTaskRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoStartTaskRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoStartTaskRequest) GetCellId() string {
	if x != nil {
		return x.CellId
	}
	return ""
}

type ProtoStartTaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error       *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	ShouldStart bool        `protobuf:"varint,2,opt,name=should_start,proto3" json:"should_start,omitempty"`
}

func (x *ProtoStartTaskResponse) Reset() {
	*x = ProtoStartTaskResponse{}
	mi := &file_task_requests_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoStartTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoStartTaskResponse) ProtoMessage() {}

func (x *ProtoStartTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoStartTaskResponse.ProtoReflect.Descriptor instead.
func (*ProtoStartTaskResponse) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{3}
}

func (x *ProtoStartTaskResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoStartTaskResponse) GetShouldStart() bool {
	if x != nil {
		return x.ShouldStart
	}
	return false
}

// Deprecated: Marked as deprecated in task_requests.proto.
type ProtoFailTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid      string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	FailureReason string `protobuf:"bytes,2,opt,name=failure_reason,proto3" json:"failure_reason,omitempty"`
}

func (x *ProtoFailTaskRequest) Reset() {
	*x = ProtoFailTaskRequest{}
	mi := &file_task_requests_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoFailTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoFailTaskRequest) ProtoMessage() {}

func (x *ProtoFailTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoFailTaskRequest.ProtoReflect.Descriptor instead.
func (*ProtoFailTaskRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{4}
}

func (x *ProtoFailTaskRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoFailTaskRequest) GetFailureReason() string {
	if x != nil {
		return x.FailureReason
	}
	return ""
}

type ProtoRejectTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid        string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	RejectionReason string `protobuf:"bytes,2,opt,name=rejection_reason,proto3" json:"rejection_reason,omitempty"`
}

func (x *ProtoRejectTaskRequest) Reset() {
	*x = ProtoRejectTaskRequest{}
	mi := &file_task_requests_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoRejectTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoRejectTaskRequest) ProtoMessage() {}

func (x *ProtoRejectTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoRejectTaskRequest.ProtoReflect.Descriptor instead.
func (*ProtoRejectTaskRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{5}
}

func (x *ProtoRejectTaskRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoRejectTaskRequest) GetRejectionReason() string {
	if x != nil {
		return x.RejectionReason
	}
	return ""
}

type ProtoTaskGuidRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
}

func (x *ProtoTaskGuidRequest) Reset() {
	*x = ProtoTaskGuidRequest{}
	mi := &file_task_requests_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTaskGuidRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTaskGuidRequest) ProtoMessage() {}

func (x *ProtoTaskGuidRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTaskGuidRequest.ProtoReflect.Descriptor instead.
func (*ProtoTaskGuidRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{6}
}

func (x *ProtoTaskGuidRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

type ProtoCompleteTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid      string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	CellId        string `protobuf:"bytes,2,opt,name=cell_id,proto3" json:"cell_id,omitempty"`
	Failed        bool   `protobuf:"varint,3,opt,name=failed,proto3" json:"failed,omitempty"`
	FailureReason string `protobuf:"bytes,4,opt,name=failure_reason,proto3" json:"failure_reason,omitempty"`
	Result        string `protobuf:"bytes,5,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *ProtoCompleteTaskRequest) Reset() {
	*x = ProtoCompleteTaskRequest{}
	mi := &file_task_requests_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoCompleteTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoCompleteTaskRequest) ProtoMessage() {}

func (x *ProtoCompleteTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoCompleteTaskRequest.ProtoReflect.Descriptor instead.
func (*ProtoCompleteTaskRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{7}
}

func (x *ProtoCompleteTaskRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoCompleteTaskRequest) GetCellId() string {
	if x != nil {
		return x.CellId
	}
	return ""
}

func (x *ProtoCompleteTaskRequest) GetFailed() bool {
	if x != nil {
		return x.Failed
	}
	return false
}

func (x *ProtoCompleteTaskRequest) GetFailureReason() string {
	if x != nil {
		return x.FailureReason
	}
	return ""
}

func (x *ProtoCompleteTaskRequest) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

type ProtoTaskCallbackResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid      string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
	Failed        bool   `protobuf:"varint,2,opt,name=failed,proto3" json:"failed,omitempty"`
	FailureReason string `protobuf:"bytes,3,opt,name=failure_reason,proto3" json:"failure_reason,omitempty"`
	Result        string `protobuf:"bytes,4,opt,name=result,proto3" json:"result,omitempty"`
	Annotation    string `protobuf:"bytes,5,opt,name=annotation,proto3" json:"annotation,omitempty"`
	CreatedAt     int64  `protobuf:"varint,6,opt,name=created_at,proto3" json:"created_at,omitempty"`
}

func (x *ProtoTaskCallbackResponse) Reset() {
	*x = ProtoTaskCallbackResponse{}
	mi := &file_task_requests_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTaskCallbackResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTaskCallbackResponse) ProtoMessage() {}

func (x *ProtoTaskCallbackResponse) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTaskCallbackResponse.ProtoReflect.Descriptor instead.
func (*ProtoTaskCallbackResponse) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{8}
}

func (x *ProtoTaskCallbackResponse) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

func (x *ProtoTaskCallbackResponse) GetFailed() bool {
	if x != nil {
		return x.Failed
	}
	return false
}

func (x *ProtoTaskCallbackResponse) GetFailureReason() string {
	if x != nil {
		return x.FailureReason
	}
	return ""
}

func (x *ProtoTaskCallbackResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

func (x *ProtoTaskCallbackResponse) GetAnnotation() string {
	if x != nil {
		return x.Annotation
	}
	return ""
}

func (x *ProtoTaskCallbackResponse) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

type ProtoTasksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domain string `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	CellId string `protobuf:"bytes,2,opt,name=cell_id,proto3" json:"cell_id,omitempty"`
}

func (x *ProtoTasksRequest) Reset() {
	*x = ProtoTasksRequest{}
	mi := &file_task_requests_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTasksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTasksRequest) ProtoMessage() {}

func (x *ProtoTasksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTasksRequest.ProtoReflect.Descriptor instead.
func (*ProtoTasksRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{9}
}

func (x *ProtoTasksRequest) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *ProtoTasksRequest) GetCellId() string {
	if x != nil {
		return x.CellId
	}
	return ""
}

type ProtoTasksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError  `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Tasks []*ProtoTask `protobuf:"bytes,2,rep,name=tasks,proto3" json:"tasks,omitempty"`
}

func (x *ProtoTasksResponse) Reset() {
	*x = ProtoTasksResponse{}
	mi := &file_task_requests_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTasksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTasksResponse) ProtoMessage() {}

func (x *ProtoTasksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTasksResponse.ProtoReflect.Descriptor instead.
func (*ProtoTasksResponse) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{10}
}

func (x *ProtoTasksResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoTasksResponse) GetTasks() []*ProtoTask {
	if x != nil {
		return x.Tasks
	}
	return nil
}

type ProtoTaskByGuidRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskGuid string `protobuf:"bytes,1,opt,name=task_guid,proto3" json:"task_guid,omitempty"`
}

func (x *ProtoTaskByGuidRequest) Reset() {
	*x = ProtoTaskByGuidRequest{}
	mi := &file_task_requests_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTaskByGuidRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTaskByGuidRequest) ProtoMessage() {}

func (x *ProtoTaskByGuidRequest) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTaskByGuidRequest.ProtoReflect.Descriptor instead.
func (*ProtoTaskByGuidRequest) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{11}
}

func (x *ProtoTaskByGuidRequest) GetTaskGuid() string {
	if x != nil {
		return x.TaskGuid
	}
	return ""
}

type ProtoTaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Task  *ProtoTask  `protobuf:"bytes,2,opt,name=task,proto3" json:"task,omitempty"`
}

func (x *ProtoTaskResponse) Reset() {
	*x = ProtoTaskResponse{}
	mi := &file_task_requests_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoTaskResponse) ProtoMessage() {}

func (x *ProtoTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_task_requests_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoTaskResponse.ProtoReflect.Descriptor instead.
func (*ProtoTaskResponse) Descriptor() ([]byte, []int) {
	return file_task_requests_proto_rawDescGZIP(), []int{12}
}

func (x *ProtoTaskResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoTaskResponse) GetTask() *ProtoTask {
	if x != nil {
		return x.Task
	}
	return nil
}

var File_task_requests_proto protoreflect.FileDescriptor

var file_task_requests_proto_rawDesc = []byte{
	0x0a, 0x13, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62,
	0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x46, 0x0a, 0x1a, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x69,
	0x66, 0x65, 0x63, 0x79, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0xa4, 0x01, 0x0a, 0x16, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x4a, 0x0a, 0x0f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x64, 0x65, 0x66,
	0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b,
	0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52,
	0x0f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x21, 0x0a, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67,
	0x75, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x22, 0x59, 0x0a, 0x15, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x09, 0x74, 0x61, 0x73,
	0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e,
	0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x07,
	0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x22, 0x6b, 0x0a, 0x16, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x27, 0x0a, 0x0c, 0x73, 0x68, 0x6f, 0x75, 0x6c, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x73, 0x68, 0x6f, 0x75,
	0x6c, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x22, 0x6a, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x46, 0x61, 0x69, 0x6c, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x21, 0x0a, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67,
	0x75, 0x69, 0x64, 0x12, 0x2b, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01,
	0x52, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e,
	0x3a, 0x02, 0x18, 0x01, 0x22, 0x6c, 0x0a, 0x16, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x65, 0x6a,
	0x65, 0x63, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21,
	0x0a, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69,
	0x64, 0x12, 0x2f, 0x0a, 0x10, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01,
	0x52, 0x10, 0x72, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x61, 0x73,
	0x6f, 0x6e, 0x22, 0x39, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x47,
	0x75, 0x69, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x09, 0x74, 0x61,
	0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x22, 0xc3, 0x01,
	0x0a, 0x18, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x09, 0x74, 0x61,
	0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x12, 0x1d, 0x0a,
	0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03,
	0xc0, 0x3e, 0x01, 0x52, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x06,
	0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x42, 0x03, 0xc0, 0x3e,
	0x01, 0x52, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x12, 0x2b, 0x0a, 0x0e, 0x66, 0x61, 0x69,
	0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f,
	0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x06, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x22, 0xea, 0x01, 0x0a, 0x19, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73,
	0x6b, 0x43, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x21, 0x0a, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f,
	0x67, 0x75, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65,
	0x64, 0x12, 0x2b, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0e,
	0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1b,
	0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03,
	0xc0, 0x3e, 0x01, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x42,
	0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x22, 0x4f, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61,
	0x69, 0x6e, 0x12, 0x1d, 0x0a, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69,
	0x64, 0x22, 0x67, 0x0a, 0x12, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x12, 0x27, 0x0a, 0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x22, 0x3b, 0x0a, 0x16, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x42, 0x79, 0x47, 0x75, 0x69, 0x64, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x74, 0x61,
	0x73, 0x6b, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x22, 0x64, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x42, 0x22, 0x5a,
	0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64,
	0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_task_requests_proto_rawDescOnce sync.Once
	file_task_requests_proto_rawDescData = file_task_requests_proto_rawDesc
)

func file_task_requests_proto_rawDescGZIP() []byte {
	file_task_requests_proto_rawDescOnce.Do(func() {
		file_task_requests_proto_rawDescData = protoimpl.X.CompressGZIP(file_task_requests_proto_rawDescData)
	})
	return file_task_requests_proto_rawDescData
}

var file_task_requests_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_task_requests_proto_goTypes = []any{
	(*ProtoTaskLifecycleResponse)(nil), // 0: models.ProtoTaskLifecycleResponse
	(*ProtoDesireTaskRequest)(nil),     // 1: models.ProtoDesireTaskRequest
	(*ProtoStartTaskRequest)(nil),      // 2: models.ProtoStartTaskRequest
	(*ProtoStartTaskResponse)(nil),     // 3: models.ProtoStartTaskResponse
	(*ProtoFailTaskRequest)(nil),       // 4: models.ProtoFailTaskRequest
	(*ProtoRejectTaskRequest)(nil),     // 5: models.ProtoRejectTaskRequest
	(*ProtoTaskGuidRequest)(nil),       // 6: models.ProtoTaskGuidRequest
	(*ProtoCompleteTaskRequest)(nil),   // 7: models.ProtoCompleteTaskRequest
	(*ProtoTaskCallbackResponse)(nil),  // 8: models.ProtoTaskCallbackResponse
	(*ProtoTasksRequest)(nil),          // 9: models.ProtoTasksRequest
	(*ProtoTasksResponse)(nil),         // 10: models.ProtoTasksResponse
	(*ProtoTaskByGuidRequest)(nil),     // 11: models.ProtoTaskByGuidRequest
	(*ProtoTaskResponse)(nil),          // 12: models.ProtoTaskResponse
	(*ProtoError)(nil),                 // 13: models.ProtoError
	(*ProtoTaskDefinition)(nil),        // 14: models.ProtoTaskDefinition
	(*ProtoTask)(nil),                  // 15: models.ProtoTask
}
var file_task_requests_proto_depIdxs = []int32{
	13, // 0: models.ProtoTaskLifecycleResponse.error:type_name -> models.ProtoError
	14, // 1: models.ProtoDesireTaskRequest.task_definition:type_name -> models.ProtoTaskDefinition
	13, // 2: models.ProtoStartTaskResponse.error:type_name -> models.ProtoError
	13, // 3: models.ProtoTasksResponse.error:type_name -> models.ProtoError
	15, // 4: models.ProtoTasksResponse.tasks:type_name -> models.ProtoTask
	13, // 5: models.ProtoTaskResponse.error:type_name -> models.ProtoError
	15, // 6: models.ProtoTaskResponse.task:type_name -> models.ProtoTask
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_task_requests_proto_init() }
func file_task_requests_proto_init() {
	if File_task_requests_proto != nil {
		return
	}
	file_bbs_proto_init()
	file_task_proto_init()
	file_error_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_task_requests_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_task_requests_proto_goTypes,
		DependencyIndexes: file_task_requests_proto_depIdxs,
		MessageInfos:      file_task_requests_proto_msgTypes,
	}.Build()
	File_task_requests_proto = out.File
	file_task_requests_proto_rawDesc = nil
	file_task_requests_proto_goTypes = nil
	file_task_requests_proto_depIdxs = nil
}
