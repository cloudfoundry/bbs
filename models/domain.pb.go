// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v4.25.6
// source: domain.proto

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

type ProtoDomainsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *ProtoError            `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Domains       []string               `protobuf:"bytes,2,rep,name=domains,proto3" json:"domains,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProtoDomainsResponse) Reset() {
	*x = ProtoDomainsResponse{}
	mi := &file_domain_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoDomainsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoDomainsResponse) ProtoMessage() {}

func (x *ProtoDomainsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_domain_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoDomainsResponse.ProtoReflect.Descriptor instead.
func (*ProtoDomainsResponse) Descriptor() ([]byte, []int) {
	return file_domain_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoDomainsResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoDomainsResponse) GetDomains() []string {
	if x != nil {
		return x.Domains
	}
	return nil
}

type ProtoUpsertDomainResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         *ProtoError            `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProtoUpsertDomainResponse) Reset() {
	*x = ProtoUpsertDomainResponse{}
	mi := &file_domain_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoUpsertDomainResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoUpsertDomainResponse) ProtoMessage() {}

func (x *ProtoUpsertDomainResponse) ProtoReflect() protoreflect.Message {
	mi := &file_domain_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoUpsertDomainResponse.ProtoReflect.Descriptor instead.
func (*ProtoUpsertDomainResponse) Descriptor() ([]byte, []int) {
	return file_domain_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoUpsertDomainResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

type ProtoUpsertDomainRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Domain        string                 `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	Ttl           uint32                 `protobuf:"varint,2,opt,name=ttl,proto3" json:"ttl,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProtoUpsertDomainRequest) Reset() {
	*x = ProtoUpsertDomainRequest{}
	mi := &file_domain_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoUpsertDomainRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoUpsertDomainRequest) ProtoMessage() {}

func (x *ProtoUpsertDomainRequest) ProtoReflect() protoreflect.Message {
	mi := &file_domain_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoUpsertDomainRequest.ProtoReflect.Descriptor instead.
func (*ProtoUpsertDomainRequest) Descriptor() ([]byte, []int) {
	return file_domain_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoUpsertDomainRequest) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *ProtoUpsertDomainRequest) GetTtl() uint32 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

var File_domain_proto protoreflect.FileDescriptor

var file_domain_proto_rawDesc = string([]byte{
	0x0a, 0x0c, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5a,
	0x0a, 0x14, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x12, 0x18, 0x0a, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0x45, 0x0a, 0x19, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x22, 0x4e, 0x0a, 0x18, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74,
	0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a,
	0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x15, 0x0a, 0x03, 0x74, 0x74,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x03, 0x74, 0x74,
	0x6c, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66,
	0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_domain_proto_rawDescOnce sync.Once
	file_domain_proto_rawDescData []byte
)

func file_domain_proto_rawDescGZIP() []byte {
	file_domain_proto_rawDescOnce.Do(func() {
		file_domain_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_domain_proto_rawDesc), len(file_domain_proto_rawDesc)))
	})
	return file_domain_proto_rawDescData
}

var file_domain_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_domain_proto_goTypes = []any{
	(*ProtoDomainsResponse)(nil),      // 0: models.ProtoDomainsResponse
	(*ProtoUpsertDomainResponse)(nil), // 1: models.ProtoUpsertDomainResponse
	(*ProtoUpsertDomainRequest)(nil),  // 2: models.ProtoUpsertDomainRequest
	(*ProtoError)(nil),                // 3: models.ProtoError
}
var file_domain_proto_depIdxs = []int32{
	3, // 0: models.ProtoDomainsResponse.error:type_name -> models.ProtoError
	3, // 1: models.ProtoUpsertDomainResponse.error:type_name -> models.ProtoError
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_domain_proto_init() }
func file_domain_proto_init() {
	if File_domain_proto != nil {
		return
	}
	file_bbs_proto_init()
	file_error_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_domain_proto_rawDesc), len(file_domain_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_domain_proto_goTypes,
		DependencyIndexes: file_domain_proto_depIdxs,
		MessageInfos:      file_domain_proto_msgTypes,
	}.Build()
	File_domain_proto = out.File
	file_domain_proto_goTypes = nil
	file_domain_proto_depIdxs = nil
}
