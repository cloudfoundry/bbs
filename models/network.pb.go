// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v4.25.6
// source: network.proto

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

type ProtoNetwork struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Properties    map[string]string      `protobuf:"bytes,1,rep,name=properties,proto3" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProtoNetwork) Reset() {
	*x = ProtoNetwork{}
	mi := &file_network_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoNetwork) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoNetwork) ProtoMessage() {}

func (x *ProtoNetwork) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoNetwork.ProtoReflect.Descriptor instead.
func (*ProtoNetwork) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoNetwork) GetProperties() map[string]string {
	if x != nil {
		return x.Properties
	}
	return nil
}

var File_network_proto protoreflect.FileDescriptor

var file_network_proto_rawDesc = string([]byte{
	0x0a, 0x0d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x22, 0x93, 0x01, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x44, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x70,
	0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x4e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x1a, 0x3d,
	0x0a, 0x0f, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x22, 0x5a,
	0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64,
	0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_network_proto_rawDescOnce sync.Once
	file_network_proto_rawDescData []byte
)

func file_network_proto_rawDescGZIP() []byte {
	file_network_proto_rawDescOnce.Do(func() {
		file_network_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_network_proto_rawDesc), len(file_network_proto_rawDesc)))
	})
	return file_network_proto_rawDescData
}

var file_network_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_network_proto_goTypes = []any{
	(*ProtoNetwork)(nil), // 0: models.ProtoNetwork
	nil,                  // 1: models.ProtoNetwork.PropertiesEntry
}
var file_network_proto_depIdxs = []int32{
	1, // 0: models.ProtoNetwork.properties:type_name -> models.ProtoNetwork.PropertiesEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_network_proto_init() }
func file_network_proto_init() {
	if File_network_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_network_proto_rawDesc), len(file_network_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_network_proto_goTypes,
		DependencyIndexes: file_network_proto_depIdxs,
		MessageInfos:      file_network_proto_msgTypes,
	}.Build()
	File_network_proto = out.File
	file_network_proto_goTypes = nil
	file_network_proto_depIdxs = nil
}
