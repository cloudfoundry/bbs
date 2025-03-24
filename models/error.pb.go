// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.4
// source: error.proto

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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          ProtoError_Type        `protobuf:"varint,1,opt,name=type,proto3,enum=models.ProtoError_Type" json:"type,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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

const file_error_proto_rawDesc = "" +
	"\n" +
	"\verror.proto\x12\x06models\x1a\tbbs.proto\"\xcc\x05\n" +
	"\n" +
	"ProtoError\x120\n" +
	"\x04type\x18\x01 \x01(\x0e2\x17.models.ProtoError.TypeB\x03\xc0>\x01R\x04type\x12 \n" +
	"\amessage\x18\x02 \x01(\tB\x06\xc0>\x01\xe0?\x01R\amessage\"\xe9\x04\n" +
	"\x04Type\x12\x10\n" +
	"\fUnknownError\x10\x00\x12\x11\n" +
	"\rInvalidRecord\x10\x03\x12\x12\n" +
	"\x0eInvalidRequest\x10\x04\x12\x13\n" +
	"\x0fInvalidResponse\x10\x05\x12\x1a\n" +
	"\x16InvalidProtobufMessage\x10\x06\x12\x0f\n" +
	"\vInvalidJSON\x10\a\x12\x18\n" +
	"\x14FailedToOpenEnvelope\x10\b\x12\x1a\n" +
	"\x16InvalidStateTransition\x10\t\x12\x14\n" +
	"\x10ResourceConflict\x10\v\x12\x12\n" +
	"\x0eResourceExists\x10\f\x12\x14\n" +
	"\x10ResourceNotFound\x10\r\x12\x0f\n" +
	"\vRouterError\x10\x0e\x12\x1c\n" +
	"\x18ActualLRPCannotBeClaimed\x10\x0f\x12\x1c\n" +
	"\x18ActualLRPCannotBeStarted\x10\x10\x12\x1c\n" +
	"\x18ActualLRPCannotBeCrashed\x10\x11\x12\x1b\n" +
	"\x17ActualLRPCannotBeFailed\x10\x12\x12\x1c\n" +
	"\x18ActualLRPCannotBeRemoved\x10\x13\x12\x1e\n" +
	"\x1aActualLRPCannotBeUnclaimed\x10\x15\x12\x1a\n" +
	"\x16RunningOnDifferentCell\x10\x18\x12\x12\n" +
	"\x0eGUIDGeneration\x10\x1a\x12\x0f\n" +
	"\vDeserialize\x10\x1b\x12\f\n" +
	"\bDeadlock\x10\x1c\x12\x11\n" +
	"\rUnrecoverable\x10\x1d\x12\x11\n" +
	"\rLockCollision\x10\x1e\x12\v\n" +
	"\aTimeout\x10\x1f\"\x04\b\x01\x10\x01\"\x04\b\x02\x10\x02\"\x04\b\n" +
	"\x10\n" +
	"\"\x04\b\x14\x10\x14\"\x04\b\x16\x10\x16\"\x04\b\x17\x10\x17\"\x04\b\x19\x10\x19B\"Z code.cloudfoundry.org/bbs/modelsb\x06proto3"

var (
	file_error_proto_rawDescOnce sync.Once
	file_error_proto_rawDescData []byte
)

func file_error_proto_rawDescGZIP() []byte {
	file_error_proto_rawDescOnce.Do(func() {
		file_error_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_error_proto_rawDesc), len(file_error_proto_rawDesc)))
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
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_error_proto_rawDesc), len(file_error_proto_rawDesc)),
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
	file_error_proto_goTypes = nil
	file_error_proto_depIdxs = nil
}
