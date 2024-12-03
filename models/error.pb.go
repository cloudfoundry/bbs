// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.0
// source: error.proto

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

type ProtoError_Type int32

const (
	ProtoError_UnknownError               ProtoError_Type = 0
	ProtoError_InvalidRecord              ProtoError_Type = 3
	ProtoError_InvalidRequest             ProtoError_Type = 4
	ProtoError_InvalidResponse            ProtoError_Type = 5
	ProtoError_InvalidProtobufMessage     ProtoError_Type = 6
	ProtoError_InvalidJSON                ProtoError_Type = 7
	ProtoError_FailedToOpenEnvelope       ProtoError_Type = 8
	ProtoError_InvalidStateTransition     ProtoError_Type = 9
	ProtoError_ResourceConflict           ProtoError_Type = 11
	ProtoError_ResourceExists             ProtoError_Type = 12
	ProtoError_ResourceNotFound           ProtoError_Type = 13
	ProtoError_RouterError                ProtoError_Type = 14
	ProtoError_ActualLRPCannotBeClaimed   ProtoError_Type = 15
	ProtoError_ActualLRPCannotBeStarted   ProtoError_Type = 16
	ProtoError_ActualLRPCannotBeCrashed   ProtoError_Type = 17
	ProtoError_ActualLRPCannotBeFailed    ProtoError_Type = 18
	ProtoError_ActualLRPCannotBeRemoved   ProtoError_Type = 19
	ProtoError_ActualLRPCannotBeUnclaimed ProtoError_Type = 21
	ProtoError_RunningOnDifferentCell     ProtoError_Type = 24
	ProtoError_GUIDGeneration             ProtoError_Type = 26
	ProtoError_Deserialize                ProtoError_Type = 27
	ProtoError_Deadlock                   ProtoError_Type = 28
	ProtoError_Unrecoverable              ProtoError_Type = 29
	ProtoError_LockCollision              ProtoError_Type = 30
	ProtoError_Timeout                    ProtoError_Type = 31
)

// Enum value maps for ProtoError_Type.
var (
	ProtoError_Type_name = map[int32]string{
		0:  "UnknownError",
		3:  "InvalidRecord",
		4:  "InvalidRequest",
		5:  "InvalidResponse",
		6:  "InvalidProtobufMessage",
		7:  "InvalidJSON",
		8:  "FailedToOpenEnvelope",
		9:  "InvalidStateTransition",
		11: "ResourceConflict",
		12: "ResourceExists",
		13: "ResourceNotFound",
		14: "RouterError",
		15: "ActualLRPCannotBeClaimed",
		16: "ActualLRPCannotBeStarted",
		17: "ActualLRPCannotBeCrashed",
		18: "ActualLRPCannotBeFailed",
		19: "ActualLRPCannotBeRemoved",
		21: "ActualLRPCannotBeUnclaimed",
		24: "RunningOnDifferentCell",
		26: "GUIDGeneration",
		27: "Deserialize",
		28: "Deadlock",
		29: "Unrecoverable",
		30: "LockCollision",
		31: "Timeout",
	}
	ProtoError_Type_value = map[string]int32{
		"UnknownError":               0,
		"InvalidRecord":              3,
		"InvalidRequest":             4,
		"InvalidResponse":            5,
		"InvalidProtobufMessage":     6,
		"InvalidJSON":                7,
		"FailedToOpenEnvelope":       8,
		"InvalidStateTransition":     9,
		"ResourceConflict":           11,
		"ResourceExists":             12,
		"ResourceNotFound":           13,
		"RouterError":                14,
		"ActualLRPCannotBeClaimed":   15,
		"ActualLRPCannotBeStarted":   16,
		"ActualLRPCannotBeCrashed":   17,
		"ActualLRPCannotBeFailed":    18,
		"ActualLRPCannotBeRemoved":   19,
		"ActualLRPCannotBeUnclaimed": 21,
		"RunningOnDifferentCell":     24,
		"GUIDGeneration":             26,
		"Deserialize":                27,
		"Deadlock":                   28,
		"Unrecoverable":              29,
		"LockCollision":              30,
		"Timeout":                    31,
	}
)

func (x ProtoError_Type) Enum() *ProtoError_Type {
	p := new(ProtoError_Type)
	*p = x
	return p
}

func (x ProtoError_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProtoError_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_error_proto_enumTypes[0].Descriptor()
}

func (ProtoError_Type) Type() protoreflect.EnumType {
	return &file_error_proto_enumTypes[0]
}

func (x ProtoError_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProtoError_Type.Descriptor instead.
func (ProtoError_Type) EnumDescriptor() ([]byte, []int) {
	return file_error_proto_rawDescGZIP(), []int{0, 0}
}

type ProtoError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type    ProtoError_Type `protobuf:"varint,1,opt,name=type,proto3,enum=models.ProtoError_Type" json:"type,omitempty"`
	Message string          `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ProtoError) Reset() {
	*x = ProtoError{}
	mi := &file_error_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoError) ProtoMessage() {}

func (x *ProtoError) ProtoReflect() protoreflect.Message {
	mi := &file_error_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoError.ProtoReflect.Descriptor instead.
func (*ProtoError) Descriptor() ([]byte, []int) {
	return file_error_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoError) GetType() ProtoError_Type {
	if x != nil {
		return x.Type
	}
	return ProtoError_UnknownError
}

func (x *ProtoError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_error_proto protoreflect.FileDescriptor

var file_error_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xcc, 0x05, 0x0a, 0x0a, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x30, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x20, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x06, 0xc0, 0x3e, 0x01, 0xe0, 0x3f, 0x01, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0xe9, 0x04, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c,
	0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x00, 0x12, 0x11,
	0x0a, 0x0d, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x10,
	0x03, 0x12, 0x12, 0x0a, 0x0e, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x10, 0x04, 0x12, 0x13, 0x0a, 0x0f, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x10, 0x05, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x6e,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x10, 0x06, 0x12, 0x0f, 0x0a, 0x0b, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x4a, 0x53, 0x4f, 0x4e, 0x10, 0x07, 0x12, 0x18, 0x0a, 0x14, 0x46, 0x61, 0x69, 0x6c, 0x65,
	0x64, 0x54, 0x6f, 0x4f, 0x70, 0x65, 0x6e, 0x45, 0x6e, 0x76, 0x65, 0x6c, 0x6f, 0x70, 0x65, 0x10,
	0x08, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x09, 0x12, 0x14, 0x0a,
	0x10, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63,
	0x74, 0x10, 0x0b, 0x12, 0x12, 0x0a, 0x0e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x45,
	0x78, 0x69, 0x73, 0x74, 0x73, 0x10, 0x0c, 0x12, 0x14, 0x0a, 0x10, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4e, 0x6f, 0x74, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x10, 0x0d, 0x12, 0x0f, 0x0a,
	0x0b, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x0e, 0x12, 0x1c,
	0x0a, 0x18, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x43, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x42, 0x65, 0x43, 0x6c, 0x61, 0x69, 0x6d, 0x65, 0x64, 0x10, 0x0f, 0x12, 0x1c, 0x0a, 0x18,
	0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x43, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x42,
	0x65, 0x53, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x10, 0x10, 0x12, 0x1c, 0x0a, 0x18, 0x41, 0x63,
	0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50, 0x43, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x42, 0x65, 0x43,
	0x72, 0x61, 0x73, 0x68, 0x65, 0x64, 0x10, 0x11, 0x12, 0x1b, 0x0a, 0x17, 0x41, 0x63, 0x74, 0x75,
	0x61, 0x6c, 0x4c, 0x52, 0x50, 0x43, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x42, 0x65, 0x46, 0x61, 0x69,
	0x6c, 0x65, 0x64, 0x10, 0x12, 0x12, 0x1c, 0x0a, 0x18, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c,
	0x52, 0x50, 0x43, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x42, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x64, 0x10, 0x13, 0x12, 0x1e, 0x0a, 0x1a, 0x41, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x4c, 0x52, 0x50,
	0x43, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x42, 0x65, 0x55, 0x6e, 0x63, 0x6c, 0x61, 0x69, 0x6d, 0x65,
	0x64, 0x10, 0x15, 0x12, 0x1a, 0x0a, 0x16, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x4f, 0x6e,
	0x44, 0x69, 0x66, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x74, 0x43, 0x65, 0x6c, 0x6c, 0x10, 0x18, 0x12,
	0x12, 0x0a, 0x0e, 0x47, 0x55, 0x49, 0x44, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x10, 0x1a, 0x12, 0x0f, 0x0a, 0x0b, 0x44, 0x65, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69,
	0x7a, 0x65, 0x10, 0x1b, 0x12, 0x0c, 0x0a, 0x08, 0x44, 0x65, 0x61, 0x64, 0x6c, 0x6f, 0x63, 0x6b,
	0x10, 0x1c, 0x12, 0x11, 0x0a, 0x0d, 0x55, 0x6e, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x61,
	0x62, 0x6c, 0x65, 0x10, 0x1d, 0x12, 0x11, 0x0a, 0x0d, 0x4c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x6c,
	0x6c, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x10, 0x1e, 0x12, 0x0b, 0x0a, 0x07, 0x54, 0x69, 0x6d, 0x65,
	0x6f, 0x75, 0x74, 0x10, 0x1f, 0x22, 0x04, 0x08, 0x01, 0x10, 0x01, 0x22, 0x04, 0x08, 0x02, 0x10,
	0x02, 0x22, 0x04, 0x08, 0x0a, 0x10, 0x0a, 0x22, 0x04, 0x08, 0x14, 0x10, 0x14, 0x22, 0x04, 0x08,
	0x16, 0x10, 0x16, 0x22, 0x04, 0x08, 0x17, 0x10, 0x17, 0x22, 0x04, 0x08, 0x19, 0x10, 0x19, 0x42,
	0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75,
	0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_error_proto_rawDescOnce sync.Once
	file_error_proto_rawDescData = file_error_proto_rawDesc
)

func file_error_proto_rawDescGZIP() []byte {
	file_error_proto_rawDescOnce.Do(func() {
		file_error_proto_rawDescData = protoimpl.X.CompressGZIP(file_error_proto_rawDescData)
	})
	return file_error_proto_rawDescData
}

var file_error_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_error_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_error_proto_goTypes = []any{
	(ProtoError_Type)(0), // 0: models.ProtoError.Type
	(*ProtoError)(nil),   // 1: models.ProtoError
}
var file_error_proto_depIdxs = []int32{
	0, // 0: models.ProtoError.type:type_name -> models.ProtoError.Type
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_error_proto_init() }
func file_error_proto_init() {
	if File_error_proto != nil {
		return
	}
	file_bbs_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_error_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_error_proto_goTypes,
		DependencyIndexes: file_error_proto_depIdxs,
		EnumInfos:         file_error_proto_enumTypes,
		MessageInfos:      file_error_proto_msgTypes,
	}.Build()
	File_error_proto = out.File
	file_error_proto_rawDesc = nil
	file_error_proto_goTypes = nil
	file_error_proto_depIdxs = nil
}
