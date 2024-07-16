// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.2
// source: evacuation.proto

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

type ProtoEvacuationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error         *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	KeepContainer bool        `protobuf:"varint,2,opt,name=keep_container,proto3" json:"keep_container,omitempty"`
}

func (x *ProtoEvacuationResponse) Reset() {
	*x = ProtoEvacuationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEvacuationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEvacuationResponse) ProtoMessage() {}

func (x *ProtoEvacuationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEvacuationResponse.ProtoReflect.Descriptor instead.
func (*ProtoEvacuationResponse) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoEvacuationResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoEvacuationResponse) GetKeepContainer() bool {
	if x != nil {
		return x.KeepContainer
	}
	return false
}

type ProtoEvacuateClaimedActualLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualLrpKey         *ProtoActualLRPKey         `protobuf:"bytes,1,opt,name=actual_lrp_key,proto3" json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ProtoActualLRPInstanceKey `protobuf:"bytes,2,opt,name=actual_lrp_instance_key,proto3" json:"actual_lrp_instance_key,omitempty"`
}

func (x *ProtoEvacuateClaimedActualLRPRequest) Reset() {
	*x = ProtoEvacuateClaimedActualLRPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEvacuateClaimedActualLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEvacuateClaimedActualLRPRequest) ProtoMessage() {}

func (x *ProtoEvacuateClaimedActualLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEvacuateClaimedActualLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoEvacuateClaimedActualLRPRequest) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoEvacuateClaimedActualLRPRequest) GetActualLrpKey() *ProtoActualLRPKey {
	if x != nil {
		return x.ActualLrpKey
	}
	return nil
}

func (x *ProtoEvacuateClaimedActualLRPRequest) GetActualLrpInstanceKey() *ProtoActualLRPInstanceKey {
	if x != nil {
		return x.ActualLrpInstanceKey
	}
	return nil
}

type ProtoEvacuateRunningActualLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualLrpKey            *ProtoActualLRPKey             `protobuf:"bytes,1,opt,name=actual_lrp_key,proto3" json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey    *ProtoActualLRPInstanceKey     `protobuf:"bytes,2,opt,name=actual_lrp_instance_key,proto3" json:"actual_lrp_instance_key,omitempty"`
	ActualLrpNetInfo        *ProtoActualLRPNetInfo         `protobuf:"bytes,3,opt,name=actual_lrp_net_info,proto3" json:"actual_lrp_net_info,omitempty"`
	ActualLrpInternalRoutes []*ProtoActualLRPInternalRoute `protobuf:"bytes,5,rep,name=actual_lrp_internal_routes,proto3" json:"actual_lrp_internal_routes,omitempty"`
	MetricTags              map[string]string              `protobuf:"bytes,6,rep,name=metric_tags,proto3" json:"metric_tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Routable                *bool                          `protobuf:"varint,7,opt,name=Routable,proto3,oneof" json:"Routable,omitempty"`
	AvailabilityZone        string                         `protobuf:"bytes,8,opt,name=availability_zone,proto3" json:"availability_zone,omitempty"`
}

func (x *ProtoEvacuateRunningActualLRPRequest) Reset() {
	*x = ProtoEvacuateRunningActualLRPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEvacuateRunningActualLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEvacuateRunningActualLRPRequest) ProtoMessage() {}

func (x *ProtoEvacuateRunningActualLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEvacuateRunningActualLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoEvacuateRunningActualLRPRequest) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetActualLrpKey() *ProtoActualLRPKey {
	if x != nil {
		return x.ActualLrpKey
	}
	return nil
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetActualLrpInstanceKey() *ProtoActualLRPInstanceKey {
	if x != nil {
		return x.ActualLrpInstanceKey
	}
	return nil
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetActualLrpNetInfo() *ProtoActualLRPNetInfo {
	if x != nil {
		return x.ActualLrpNetInfo
	}
	return nil
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetActualLrpInternalRoutes() []*ProtoActualLRPInternalRoute {
	if x != nil {
		return x.ActualLrpInternalRoutes
	}
	return nil
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetMetricTags() map[string]string {
	if x != nil {
		return x.MetricTags
	}
	return nil
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetRoutable() bool {
	if x != nil && x.Routable != nil {
		return *x.Routable
	}
	return false
}

func (x *ProtoEvacuateRunningActualLRPRequest) GetAvailabilityZone() string {
	if x != nil {
		return x.AvailabilityZone
	}
	return ""
}

type ProtoEvacuateStoppedActualLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualLrpKey         *ProtoActualLRPKey         `protobuf:"bytes,1,opt,name=actual_lrp_key,proto3" json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ProtoActualLRPInstanceKey `protobuf:"bytes,2,opt,name=actual_lrp_instance_key,proto3" json:"actual_lrp_instance_key,omitempty"`
}

func (x *ProtoEvacuateStoppedActualLRPRequest) Reset() {
	*x = ProtoEvacuateStoppedActualLRPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEvacuateStoppedActualLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEvacuateStoppedActualLRPRequest) ProtoMessage() {}

func (x *ProtoEvacuateStoppedActualLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEvacuateStoppedActualLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoEvacuateStoppedActualLRPRequest) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{3}
}

func (x *ProtoEvacuateStoppedActualLRPRequest) GetActualLrpKey() *ProtoActualLRPKey {
	if x != nil {
		return x.ActualLrpKey
	}
	return nil
}

func (x *ProtoEvacuateStoppedActualLRPRequest) GetActualLrpInstanceKey() *ProtoActualLRPInstanceKey {
	if x != nil {
		return x.ActualLrpInstanceKey
	}
	return nil
}

type ProtoEvacuateCrashedActualLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualLrpKey         *ProtoActualLRPKey         `protobuf:"bytes,1,opt,name=actual_lrp_key,proto3" json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ProtoActualLRPInstanceKey `protobuf:"bytes,2,opt,name=actual_lrp_instance_key,proto3" json:"actual_lrp_instance_key,omitempty"`
	ErrorMessage         string                     `protobuf:"bytes,3,opt,name=error_message,proto3" json:"error_message,omitempty"`
}

func (x *ProtoEvacuateCrashedActualLRPRequest) Reset() {
	*x = ProtoEvacuateCrashedActualLRPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEvacuateCrashedActualLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEvacuateCrashedActualLRPRequest) ProtoMessage() {}

func (x *ProtoEvacuateCrashedActualLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEvacuateCrashedActualLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoEvacuateCrashedActualLRPRequest) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{4}
}

func (x *ProtoEvacuateCrashedActualLRPRequest) GetActualLrpKey() *ProtoActualLRPKey {
	if x != nil {
		return x.ActualLrpKey
	}
	return nil
}

func (x *ProtoEvacuateCrashedActualLRPRequest) GetActualLrpInstanceKey() *ProtoActualLRPInstanceKey {
	if x != nil {
		return x.ActualLrpInstanceKey
	}
	return nil
}

func (x *ProtoEvacuateCrashedActualLRPRequest) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

type ProtoRemoveEvacuatingActualLRPRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActualLrpKey         *ProtoActualLRPKey         `protobuf:"bytes,1,opt,name=actual_lrp_key,proto3" json:"actual_lrp_key,omitempty"`
	ActualLrpInstanceKey *ProtoActualLRPInstanceKey `protobuf:"bytes,2,opt,name=actual_lrp_instance_key,proto3" json:"actual_lrp_instance_key,omitempty"`
}

func (x *ProtoRemoveEvacuatingActualLRPRequest) Reset() {
	*x = ProtoRemoveEvacuatingActualLRPRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoRemoveEvacuatingActualLRPRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoRemoveEvacuatingActualLRPRequest) ProtoMessage() {}

func (x *ProtoRemoveEvacuatingActualLRPRequest) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoRemoveEvacuatingActualLRPRequest.ProtoReflect.Descriptor instead.
func (*ProtoRemoveEvacuatingActualLRPRequest) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{5}
}

func (x *ProtoRemoveEvacuatingActualLRPRequest) GetActualLrpKey() *ProtoActualLRPKey {
	if x != nil {
		return x.ActualLrpKey
	}
	return nil
}

func (x *ProtoRemoveEvacuatingActualLRPRequest) GetActualLrpInstanceKey() *ProtoActualLRPInstanceKey {
	if x != nil {
		return x.ActualLrpInstanceKey
	}
	return nil
}

type ProtoRemoveEvacuatingActualLRPResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ProtoRemoveEvacuatingActualLRPResponse) Reset() {
	*x = ProtoRemoveEvacuatingActualLRPResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_evacuation_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoRemoveEvacuatingActualLRPResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoRemoveEvacuatingActualLRPResponse) ProtoMessage() {}

func (x *ProtoRemoveEvacuatingActualLRPResponse) ProtoReflect() protoreflect.Message {
	mi := &file_evacuation_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoRemoveEvacuatingActualLRPResponse.ProtoReflect.Descriptor instead.
func (*ProtoRemoveEvacuatingActualLRPResponse) Descriptor() ([]byte, []int) {
	return file_evacuation_proto_rawDescGZIP(), []int{6}
}

func (x *ProtoRemoveEvacuatingActualLRPResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

var File_evacuation_proto protoreflect.FileDescriptor

var file_evacuation_proto_rawDesc = []byte{
	0x0a, 0x10, 0x65, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x70, 0x0a, 0x17, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x76, 0x61,
	0x63, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x2b, 0x0a, 0x0e, 0x6b, 0x65, 0x65,
	0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0e, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x22, 0xc6, 0x01, 0x0a, 0x24, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x45, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74, 0x65, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x65, 0x64, 0x41,
	0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x41, 0x0a, 0x0e, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4b,
	0x65, 0x79, 0x52, 0x0e, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b,
	0x65, 0x79, 0x12, 0x5b, 0x0a, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70,
	0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x4b, 0x65, 0x79, 0x52, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c,
	0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x22,
	0x82, 0x05, 0x0a, 0x24, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74,
	0x65, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52,
	0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x0e, 0x61, 0x63, 0x74, 0x75,
	0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41,
	0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4b, 0x65, 0x79, 0x52, 0x0e, 0x61, 0x63, 0x74,
	0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x5b, 0x0a, 0x17, 0x61,
	0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x4c, 0x52, 0x50, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4b, 0x65, 0x79, 0x52,
	0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x4f, 0x0a, 0x13, 0x61, 0x63, 0x74, 0x75,
	0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6e, 0x65, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4e, 0x65, 0x74,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x13, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70,
	0x5f, 0x6e, 0x65, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x12, 0x63, 0x0a, 0x1a, 0x61, 0x63, 0x74,
	0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75,
	0x61, 0x6c, 0x4c, 0x52, 0x50, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x52, 0x6f, 0x75,
	0x74, 0x65, 0x52, 0x1a, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x12, 0x5e,
	0x0a, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x3c, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x45, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e,
	0x67, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x12, 0x1f,
	0x0a, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08,
	0x48, 0x00, 0x52, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x88, 0x01, 0x01, 0x12,
	0x31, 0x0a, 0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f,
	0x7a, 0x6f, 0x6e, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52,
	0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x7a, 0x6f,
	0x6e, 0x65, 0x1a, 0x3d, 0x0a, 0x0f, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x52, 0x6f, 0x75, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4a, 0x04,
	0x08, 0x04, 0x10, 0x05, 0x22, 0xc6, 0x01, 0x0a, 0x24, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x76,
	0x61, 0x63, 0x75, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x41, 0x63, 0x74,
	0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a,
	0x0e, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4b, 0x65, 0x79,
	0x52, 0x0e, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79,
	0x12, 0x5b, 0x0a, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x21, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x4b, 0x65, 0x79, 0x52, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70,
	0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x22, 0xf1, 0x01,
	0x0a, 0x24, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74, 0x65, 0x43,
	0x72, 0x61, 0x73, 0x68, 0x65, 0x64, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x0e, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c,
	0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74,
	0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4b, 0x65, 0x79, 0x52, 0x0e, 0x61, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x5b, 0x0a, 0x17, 0x61, 0x63, 0x74,
	0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c,
	0x52, 0x50, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4b, 0x65, 0x79, 0x52, 0x17, 0x61,
	0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x29, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0xc7, 0x01, 0x0a, 0x25, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x45, 0x76, 0x61, 0x63, 0x75, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x41, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x0e, 0x61,
	0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x4b, 0x65, 0x79, 0x52, 0x0e,
	0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x5b,
	0x0a, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x21, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x41, 0x63,
	0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4b,
	0x65, 0x79, 0x52, 0x17, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x5f, 0x6c, 0x72, 0x70, 0x5f, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x22, 0x52, 0x0a, 0x26, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x45, 0x76, 0x61, 0x63, 0x75, 0x61,
	0x74, 0x69, 0x6e, 0x67, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42,
	0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75,
	0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_evacuation_proto_rawDescOnce sync.Once
	file_evacuation_proto_rawDescData = file_evacuation_proto_rawDesc
)

func file_evacuation_proto_rawDescGZIP() []byte {
	file_evacuation_proto_rawDescOnce.Do(func() {
		file_evacuation_proto_rawDescData = protoimpl.X.CompressGZIP(file_evacuation_proto_rawDescData)
	})
	return file_evacuation_proto_rawDescData
}

var file_evacuation_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_evacuation_proto_goTypes = []any{
	(*ProtoEvacuationResponse)(nil),                // 0: models.ProtoEvacuationResponse
	(*ProtoEvacuateClaimedActualLRPRequest)(nil),   // 1: models.ProtoEvacuateClaimedActualLRPRequest
	(*ProtoEvacuateRunningActualLRPRequest)(nil),   // 2: models.ProtoEvacuateRunningActualLRPRequest
	(*ProtoEvacuateStoppedActualLRPRequest)(nil),   // 3: models.ProtoEvacuateStoppedActualLRPRequest
	(*ProtoEvacuateCrashedActualLRPRequest)(nil),   // 4: models.ProtoEvacuateCrashedActualLRPRequest
	(*ProtoRemoveEvacuatingActualLRPRequest)(nil),  // 5: models.ProtoRemoveEvacuatingActualLRPRequest
	(*ProtoRemoveEvacuatingActualLRPResponse)(nil), // 6: models.ProtoRemoveEvacuatingActualLRPResponse
	nil,                                 // 7: models.ProtoEvacuateRunningActualLRPRequest.MetricTagsEntry
	(*ProtoError)(nil),                  // 8: models.ProtoError
	(*ProtoActualLRPKey)(nil),           // 9: models.ProtoActualLRPKey
	(*ProtoActualLRPInstanceKey)(nil),   // 10: models.ProtoActualLRPInstanceKey
	(*ProtoActualLRPNetInfo)(nil),       // 11: models.ProtoActualLRPNetInfo
	(*ProtoActualLRPInternalRoute)(nil), // 12: models.ProtoActualLRPInternalRoute
}
var file_evacuation_proto_depIdxs = []int32{
	8,  // 0: models.ProtoEvacuationResponse.error:type_name -> models.ProtoError
	9,  // 1: models.ProtoEvacuateClaimedActualLRPRequest.actual_lrp_key:type_name -> models.ProtoActualLRPKey
	10, // 2: models.ProtoEvacuateClaimedActualLRPRequest.actual_lrp_instance_key:type_name -> models.ProtoActualLRPInstanceKey
	9,  // 3: models.ProtoEvacuateRunningActualLRPRequest.actual_lrp_key:type_name -> models.ProtoActualLRPKey
	10, // 4: models.ProtoEvacuateRunningActualLRPRequest.actual_lrp_instance_key:type_name -> models.ProtoActualLRPInstanceKey
	11, // 5: models.ProtoEvacuateRunningActualLRPRequest.actual_lrp_net_info:type_name -> models.ProtoActualLRPNetInfo
	12, // 6: models.ProtoEvacuateRunningActualLRPRequest.actual_lrp_internal_routes:type_name -> models.ProtoActualLRPInternalRoute
	7,  // 7: models.ProtoEvacuateRunningActualLRPRequest.metric_tags:type_name -> models.ProtoEvacuateRunningActualLRPRequest.MetricTagsEntry
	9,  // 8: models.ProtoEvacuateStoppedActualLRPRequest.actual_lrp_key:type_name -> models.ProtoActualLRPKey
	10, // 9: models.ProtoEvacuateStoppedActualLRPRequest.actual_lrp_instance_key:type_name -> models.ProtoActualLRPInstanceKey
	9,  // 10: models.ProtoEvacuateCrashedActualLRPRequest.actual_lrp_key:type_name -> models.ProtoActualLRPKey
	10, // 11: models.ProtoEvacuateCrashedActualLRPRequest.actual_lrp_instance_key:type_name -> models.ProtoActualLRPInstanceKey
	9,  // 12: models.ProtoRemoveEvacuatingActualLRPRequest.actual_lrp_key:type_name -> models.ProtoActualLRPKey
	10, // 13: models.ProtoRemoveEvacuatingActualLRPRequest.actual_lrp_instance_key:type_name -> models.ProtoActualLRPInstanceKey
	8,  // 14: models.ProtoRemoveEvacuatingActualLRPResponse.error:type_name -> models.ProtoError
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_evacuation_proto_init() }
func file_evacuation_proto_init() {
	if File_evacuation_proto != nil {
		return
	}
	file_bbs_proto_init()
	file_actual_lrp_proto_init()
	file_error_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_evacuation_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoEvacuationResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoEvacuateClaimedActualLRPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoEvacuateRunningActualLRPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoEvacuateStoppedActualLRPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoEvacuateCrashedActualLRPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoRemoveEvacuatingActualLRPRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_evacuation_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*ProtoRemoveEvacuatingActualLRPResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_evacuation_proto_msgTypes[2].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_evacuation_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_evacuation_proto_goTypes,
		DependencyIndexes: file_evacuation_proto_depIdxs,
		MessageInfos:      file_evacuation_proto_msgTypes,
	}.Build()
	File_evacuation_proto = out.File
	file_evacuation_proto_rawDesc = nil
	file_evacuation_proto_goTypes = nil
	file_evacuation_proto_depIdxs = nil
}
