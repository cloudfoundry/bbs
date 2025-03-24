// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.4
// source: desired_lrp_requests.proto

package models

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ProtoDesiredLRPLifecycleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *ProtoError            `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *ProtoError            `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrps   []*ProtoDesiredLRP     `protobuf:"bytes,2,rep,name=desired_lrps,proto3" json:"desired_lrps,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domain        string                 `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	ProcessGuids  []string               `protobuf:"bytes,2,rep,name=process_guids,proto3" json:"process_guids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *ProtoError            `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrp    *ProtoDesiredLRP       `protobuf:"bytes,2,opt,name=desired_lrp,proto3" json:"desired_lrp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state                     protoimpl.MessageState           `protogen:"open.v1"`
	Error                     *ProtoError                      `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrpSchedulingInfos []*ProtoDesiredLRPSchedulingInfo `protobuf:"bytes,2,rep,name=desired_lrp_scheduling_infos,proto3" json:"desired_lrp_scheduling_infos,omitempty"`
	unknownFields             protoimpl.UnknownFields
	sizeCache                 protoimpl.SizeCache
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
	state                    protoimpl.MessageState         `protogen:"open.v1"`
	Error                    *ProtoError                    `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	DesiredLrpSchedulingInfo *ProtoDesiredLRPSchedulingInfo `protobuf:"bytes,2,opt,name=desired_lrp_scheduling_info,proto3" json:"desired_lrp_scheduling_info,omitempty"`
	unknownFields            protoimpl.UnknownFields
	sizeCache                protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProcessGuid   string                 `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	DesiredLrp    *ProtoDesiredLRP       `protobuf:"bytes,1,opt,name=desired_lrp,proto3" json:"desired_lrp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProcessGuid   string                 `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
	Update        *ProtoDesiredLRPUpdate `protobuf:"bytes,2,opt,name=update,proto3" json:"update,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProcessGuid   string                 `protobuf:"bytes,1,opt,name=process_guid,proto3" json:"process_guid,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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

const file_desired_lrp_requests_proto_rawDesc = "" +
	"\n" +
	"\x1adesired_lrp_requests.proto\x12\x06models\x1a\tbbs.proto\x1a\x11desired_lrp.proto\x1a\verror.proto\"L\n" +
	" ProtoDesiredLRPLifecycleResponse\x12(\n" +
	"\x05error\x18\x01 \x01(\v2\x12.models.ProtoErrorR\x05error\"\x81\x01\n" +
	"\x18ProtoDesiredLRPsResponse\x12(\n" +
	"\x05error\x18\x01 \x01(\v2\x12.models.ProtoErrorR\x05error\x12;\n" +
	"\fdesired_lrps\x18\x02 \x03(\v2\x17.models.ProtoDesiredLRPR\fdesired_lrps\"\\\n" +
	"\x17ProtoDesiredLRPsRequest\x12\x1b\n" +
	"\x06domain\x18\x01 \x01(\tB\x03\xc0>\x01R\x06domain\x12$\n" +
	"\rprocess_guids\x18\x02 \x03(\tR\rprocess_guids\"~\n" +
	"\x17ProtoDesiredLRPResponse\x12(\n" +
	"\x05error\x18\x01 \x01(\v2\x12.models.ProtoErrorR\x05error\x129\n" +
	"\vdesired_lrp\x18\x02 \x01(\v2\x17.models.ProtoDesiredLRPR\vdesired_lrp\"\xbd\x01\n" +
	"&ProtoDesiredLRPSchedulingInfosResponse\x12(\n" +
	"\x05error\x18\x01 \x01(\v2\x12.models.ProtoErrorR\x05error\x12i\n" +
	"\x1cdesired_lrp_scheduling_infos\x18\x02 \x03(\v2%.models.ProtoDesiredLRPSchedulingInfoR\x1cdesired_lrp_scheduling_infos\"\xc7\x01\n" +
	"2ProtoDesiredLRPSchedulingInfoByProcessGuidResponse\x12(\n" +
	"\x05error\x18\x01 \x01(\v2\x12.models.ProtoErrorR\x05error\x12g\n" +
	"\x1bdesired_lrp_scheduling_info\x18\x02 \x01(\v2%.models.ProtoDesiredLRPSchedulingInfoR\x1bdesired_lrp_scheduling_info\"N\n" +
	"#ProtoDesiredLRPByProcessGuidRequest\x12'\n" +
	"\fprocess_guid\x18\x01 \x01(\tB\x03\xc0>\x01R\fprocess_guid\"R\n" +
	"\x15ProtoDesireLRPRequest\x129\n" +
	"\vdesired_lrp\x18\x01 \x01(\v2\x17.models.ProtoDesiredLRPR\vdesired_lrp\"~\n" +
	"\x1cProtoUpdateDesiredLRPRequest\x12'\n" +
	"\fprocess_guid\x18\x01 \x01(\tB\x03\xc0>\x01R\fprocess_guid\x125\n" +
	"\x06update\x18\x02 \x01(\v2\x1d.models.ProtoDesiredLRPUpdateR\x06update\"G\n" +
	"\x1cProtoRemoveDesiredLRPRequest\x12'\n" +
	"\fprocess_guid\x18\x01 \x01(\tB\x03\xc0>\x01R\fprocess_guidB\"Z code.cloudfoundry.org/bbs/modelsb\x06proto3"

var (
	file_desired_lrp_requests_proto_rawDescOnce sync.Once
	file_desired_lrp_requests_proto_rawDescData []byte
)

func file_desired_lrp_requests_proto_rawDescGZIP() []byte {
	file_desired_lrp_requests_proto_rawDescOnce.Do(func() {
		file_desired_lrp_requests_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_desired_lrp_requests_proto_rawDesc), len(file_desired_lrp_requests_proto_rawDesc)))
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
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_desired_lrp_requests_proto_rawDesc), len(file_desired_lrp_requests_proto_rawDesc)),
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
	file_desired_lrp_requests_proto_goTypes = nil
	file_desired_lrp_requests_proto_depIdxs = nil
}
