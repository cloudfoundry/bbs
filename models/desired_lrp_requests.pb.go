// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: desired_lrp_requests.proto

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

type ProtoDesiredLRPLifecycleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ProtoDesiredLRPLifecycleResponse) Reset() {
	*x = ProtoDesiredLRPLifecycleResponse{}
	mi := &file_desired_lrp_requests_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPLifecycleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPLifecycleResponse) ProtoMessage() {}

func (x *ProtoDesiredLRPLifecycleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPLifecycleResponse.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPLifecycleResponse) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoDesiredLRPLifecycleResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

type ProtoDesiredLRPsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error       *ProtoError        `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrps []*ProtoDesiredLRP `protobuf:"bytes,2,rep,name=desired_lrps,proto3" json:"desired_lrps,omitempty"`
}

func (x *ProtoDesiredLRPsResponse) Reset() {
	*x = ProtoDesiredLRPsResponse{}
	mi := &file_desired_lrp_requests_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPsResponse) ProtoMessage() {}

func (x *ProtoDesiredLRPsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPsResponse.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPsResponse) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoDesiredLRPsResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoDesiredLRPsResponse) GetDesiredLrps() []*ProtoDesiredLRP {
	if x != nil {
		return x.DesiredLrps
	}
	return nil
}

type ProtoDesiredLRPsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domain       string   `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	ProcessGuids []string `protobuf:"bytes,2,rep,name=process_guids,proto3" json:"process_guids,omitempty"`
}

func (x *ProtoDesiredLRPsRequest) Reset() {
	*x = ProtoDesiredLRPsRequest{}
	mi := &file_desired_lrp_requests_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPsRequest) ProtoMessage() {}

func (x *ProtoDesiredLRPsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPsRequest.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPsRequest) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoDesiredLRPsRequest) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *ProtoDesiredLRPsRequest) GetProcessGuids() []string {
	if x != nil {
		return x.ProcessGuids
	}
	return nil
}

type ProtoDesiredLRPResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error      *ProtoError      `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrp *ProtoDesiredLRP `protobuf:"bytes,2,opt,name=desired_lrp,proto3" json:"desired_lrp,omitempty"`
}

func (x *ProtoDesiredLRPResponse) Reset() {
	*x = ProtoDesiredLRPResponse{}
	mi := &file_desired_lrp_requests_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPResponse) ProtoMessage() {}

func (x *ProtoDesiredLRPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPResponse.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPResponse) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{3}
}

func (x *ProtoDesiredLRPResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoDesiredLRPResponse) GetDesiredLrp() *ProtoDesiredLRP {
	if x != nil {
		return x.DesiredLrp
	}
	return nil
}

type ProtoDesiredLRPSchedulingInfosResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error                     *ProtoError                      `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrpSchedulingInfos []*ProtoDesiredLRPSchedulingInfo `protobuf:"bytes,2,rep,name=desired_lrp_scheduling_infos,proto3" json:"desired_lrp_scheduling_infos,omitempty"`
}

func (x *ProtoDesiredLRPSchedulingInfosResponse) Reset() {
	*x = ProtoDesiredLRPSchedulingInfosResponse{}
	mi := &file_desired_lrp_requests_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPSchedulingInfosResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPSchedulingInfosResponse) ProtoMessage() {}

func (x *ProtoDesiredLRPSchedulingInfosResponse) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPSchedulingInfosResponse.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPSchedulingInfosResponse) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{4}
}

func (x *ProtoDesiredLRPSchedulingInfosResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoDesiredLRPSchedulingInfosResponse) GetDesiredLrpSchedulingInfos() []*ProtoDesiredLRPSchedulingInfo {
	if x != nil {
		return x.DesiredLrpSchedulingInfos
	}
	return nil
}

type ProtoDesiredLRPSchedulingInfoByProcessGuidResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error                    *ProtoError                    `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrpSchedulingInfo *ProtoDesiredLRPSchedulingInfo `protobuf:"bytes,2,opt,name=desired_lrp_scheduling_info,proto3" json:"desired_lrp_scheduling_info,omitempty"`
}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) Reset() {
	*x = ProtoDesiredLRPSchedulingInfoByProcessGuidResponse{}
	mi := &file_desired_lrp_requests_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) ProtoMessage() {}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPSchedulingInfoByProcessGuidResponse.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{5}
}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoDesiredLRPSchedulingInfoByProcessGuidResponse) GetDesiredLrpSchedulingInfo() *ProtoDesiredLRPSchedulingInfo {
	if x != nil {
		return x.DesiredLrpSchedulingInfo
	}
	return nil
}

type ProtoDesiredLRPByProcessGuidRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProcessGuid string `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
}

func (x *ProtoDesiredLRPByProcessGuidRequest) Reset() {
	*x = ProtoDesiredLRPByProcessGuidRequest{}
	mi := &file_desired_lrp_requests_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesiredLRPByProcessGuidRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesiredLRPByProcessGuidRequest) ProtoMessage() {}

func (x *ProtoDesiredLRPByProcessGuidRequest) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesiredLRPByProcessGuidRequest.ProtoReflect.Descriptor instead.
func (*ProtoDesiredLRPByProcessGuidRequest) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{6}
}

func (x *ProtoDesiredLRPByProcessGuidRequest) GetProcessGuid() string {
	if x != nil {
		return x.ProcessGuid
	}
	return ""
}

type ProtoDesireLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DesiredLrp *ProtoDesiredLRP `protobuf:"bytes,1,opt,name=desired_lrp,proto3" json:"desired_lrp,omitempty"`
}

func (x *ProtoDesireLRPRequest) Reset() {
	*x = ProtoDesireLRPRequest{}
	mi := &file_desired_lrp_requests_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDesireLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDesireLRPRequest) ProtoMessage() {}

func (x *ProtoDesireLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDesireLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoDesireLRPRequest) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{7}
}

func (x *ProtoDesireLRPRequest) GetDesiredLrp() *ProtoDesiredLRP {
	if x != nil {
		return x.DesiredLrp
	}
	return nil
}

type ProtoUpdateDesiredLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProcessGuid string                 `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
	Update      *ProtoDesiredLRPUpdate `protobuf:"bytes,2,opt,name=update,proto3" json:"update,omitempty"`
}

func (x *ProtoUpdateDesiredLRPRequest) Reset() {
	*x = ProtoUpdateDesiredLRPRequest{}
	mi := &file_desired_lrp_requests_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoUpdateDesiredLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoUpdateDesiredLRPRequest) ProtoMessage() {}

func (x *ProtoUpdateDesiredLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoUpdateDesiredLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoUpdateDesiredLRPRequest) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{8}
}

func (x *ProtoUpdateDesiredLRPRequest) GetProcessGuid() string {
	if x != nil {
		return x.ProcessGuid
	}
	return ""
}

func (x *ProtoUpdateDesiredLRPRequest) GetUpdate() *ProtoDesiredLRPUpdate {
	if x != nil {
		return x.Update
	}
	return nil
}

type ProtoRemoveDesiredLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProcessGuid string `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
}

func (x *ProtoRemoveDesiredLRPRequest) Reset() {
	*x = ProtoRemoveDesiredLRPRequest{}
	mi := &file_desired_lrp_requests_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoRemoveDesiredLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoRemoveDesiredLRPRequest) ProtoMessage() {}

func (x *ProtoRemoveDesiredLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_desired_lrp_requests_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoRemoveDesiredLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoRemoveDesiredLRPRequest) Descriptor() ([]byte, []int) {
	return file_desired_lrp_requests_proto_rawDescGZIP(), []int{9}
}

func (x *ProtoRemoveDesiredLRPRequest) GetProcessGuid() string {
	if x != nil {
		return x.ProcessGuid
	}
	return ""
}

var File_desired_lrp_requests_proto protoreflect.FileDescriptor

var file_desired_lrp_requests_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x11, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x4c, 0x0a, 0x20, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c,
	0x52, 0x50, 0x4c, 0x69, 0x66, 0x65, 0x63, 0x79, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x81, 0x01,
	0x0a, 0x18, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c, 0x52,
	0x50, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x3b, 0x0a, 0x0c, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f,
	0x6c, 0x72, 0x70, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64,
	0x4c, 0x52, 0x50, 0x52, 0x0c, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70,
	0x73, 0x22, 0x5c, 0x0a, 0x17, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65,
	0x64, 0x4c, 0x52, 0x50, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x06,
	0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e,
	0x01, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x24, 0x0a, 0x0d, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x0d, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x73, 0x22,
	0x7e, 0x0a, 0x17, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c,
	0x52, 0x50, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x39, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f,
	0x6c, 0x72, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c,
	0x52, 0x50, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x22,
	0xbd, 0x01, 0x0a, 0x26, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64,
	0x4c, 0x52, 0x50, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x66,
	0x6f, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x69, 0x0a, 0x1c, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f,
	0x6c, 0x72, 0x70, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69,
	0x6e, 0x66, 0x6f, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64,
	0x4c, 0x52, 0x50, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x1c, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x73,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x73, 0x22,
	0xc7, 0x01, 0x0a, 0x32, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64,
	0x4c, 0x52, 0x50, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x66,
	0x6f, 0x42, 0x79, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x47, 0x75, 0x69, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x12, 0x67, 0x0a, 0x1b, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x5f,
	0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c, 0x52, 0x50, 0x53, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x1b, 0x64, 0x65,
	0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75,
	0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x22, 0x4e, 0x0a, 0x23, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c, 0x52, 0x50, 0x42, 0x79, 0x50, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x47, 0x75, 0x69, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x27, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x22, 0x52, 0x0a, 0x15, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x39, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72,
	0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c, 0x52, 0x50,
	0x52, 0x0b, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x6c, 0x72, 0x70, 0x22, 0x7e, 0x0a,
	0x1c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x73, 0x69,
	0x72, 0x65, 0x64, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a,
	0x0c, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73,
	0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x12, 0x35, 0x0a, 0x06, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x4c, 0x52, 0x50, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x06, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x22, 0x47, 0x0a,
	0x1c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x44, 0x65, 0x73, 0x69,
	0x72, 0x65, 0x64, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a,
	0x0c, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73,
	0x73, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f,
	0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_desired_lrp_requests_proto_rawDescOnce sync.Once
	file_desired_lrp_requests_proto_rawDescData = file_desired_lrp_requests_proto_rawDesc
)

func file_desired_lrp_requests_proto_rawDescGZIP() []byte {
	file_desired_lrp_requests_proto_rawDescOnce.Do(func() {
		file_desired_lrp_requests_proto_rawDescData = protoimpl.X.CompressGZIP(file_desired_lrp_requests_proto_rawDescData)
	})
	return file_desired_lrp_requests_proto_rawDescData
}

var file_desired_lrp_requests_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_desired_lrp_requests_proto_goTypes = []any{
	(*ProtoDesiredLRPLifecycleResponse)(nil),                   // 0: models.ProtoDesiredLRPLifecycleResponse
	(*ProtoDesiredLRPsResponse)(nil),                           // 1: models.ProtoDesiredLRPsResponse
	(*ProtoDesiredLRPsRequest)(nil),                            // 2: models.ProtoDesiredLRPsRequest
	(*ProtoDesiredLRPResponse)(nil),                            // 3: models.ProtoDesiredLRPResponse
	(*ProtoDesiredLRPSchedulingInfosResponse)(nil),             // 4: models.ProtoDesiredLRPSchedulingInfosResponse
	(*ProtoDesiredLRPSchedulingInfoByProcessGuidResponse)(nil), // 5: models.ProtoDesiredLRPSchedulingInfoByProcessGuidResponse
	(*ProtoDesiredLRPByProcessGuidRequest)(nil),                // 6: models.ProtoDesiredLRPByProcessGuidRequest
	(*ProtoDesireLRPRequest)(nil),                              // 7: models.ProtoDesireLRPRequest
	(*ProtoUpdateDesiredLRPRequest)(nil),                       // 8: models.ProtoUpdateDesiredLRPRequest
	(*ProtoRemoveDesiredLRPRequest)(nil),                       // 9: models.ProtoRemoveDesiredLRPRequest
	(*ProtoError)(nil),                                         // 10: models.ProtoError
	(*ProtoDesiredLRP)(nil),                                    // 11: models.ProtoDesiredLRP
	(*ProtoDesiredLRPSchedulingInfo)(nil),                      // 12: models.ProtoDesiredLRPSchedulingInfo
	(*ProtoDesiredLRPUpdate)(nil),                              // 13: models.ProtoDesiredLRPUpdate
}
var file_desired_lrp_requests_proto_depIdxs = []int32{
	10, // 0: models.ProtoDesiredLRPLifecycleResponse.error:type_name -> models.ProtoError
	10, // 1: models.ProtoDesiredLRPsResponse.error:type_name -> models.ProtoError
	11, // 2: models.ProtoDesiredLRPsResponse.desired_lrps:type_name -> models.ProtoDesiredLRP
	10, // 3: models.ProtoDesiredLRPResponse.error:type_name -> models.ProtoError
	11, // 4: models.ProtoDesiredLRPResponse.desired_lrp:type_name -> models.ProtoDesiredLRP
	10, // 5: models.ProtoDesiredLRPSchedulingInfosResponse.error:type_name -> models.ProtoError
	12, // 6: models.ProtoDesiredLRPSchedulingInfosResponse.desired_lrp_scheduling_infos:type_name -> models.ProtoDesiredLRPSchedulingInfo
	10, // 7: models.ProtoDesiredLRPSchedulingInfoByProcessGuidResponse.error:type_name -> models.ProtoError
	12, // 8: models.ProtoDesiredLRPSchedulingInfoByProcessGuidResponse.desired_lrp_scheduling_info:type_name -> models.ProtoDesiredLRPSchedulingInfo
	11, // 9: models.ProtoDesireLRPRequest.desired_lrp:type_name -> models.ProtoDesiredLRP
	13, // 10: models.ProtoUpdateDesiredLRPRequest.update:type_name -> models.ProtoDesiredLRPUpdate
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_desired_lrp_requests_proto_init() }
func file_desired_lrp_requests_proto_init() {
	if File_desired_lrp_requests_proto != nil {
		return
	}
	file_bbs_proto_init()
	file_desired_lrp_proto_init()
	file_error_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_desired_lrp_requests_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_desired_lrp_requests_proto_goTypes,
		DependencyIndexes: file_desired_lrp_requests_proto_depIdxs,
		MessageInfos:      file_desired_lrp_requests_proto_msgTypes,
	}.Build()
	File_desired_lrp_requests_proto = out.File
	file_desired_lrp_requests_proto_rawDesc = nil
	file_desired_lrp_requests_proto_goTypes = nil
	file_desired_lrp_requests_proto_depIdxs = nil
}
