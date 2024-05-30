// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: metric_tags.proto

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

type ProtoMetricTagValue_DynamicValue int32

const (
	ProtoMetricTagValue_DynamicValueInvalid ProtoMetricTagValue_DynamicValue = 0
	ProtoMetricTagValue_INDEX               ProtoMetricTagValue_DynamicValue = 1
	ProtoMetricTagValue_INSTANCE_GUID       ProtoMetricTagValue_DynamicValue = 2
)

// Enum value maps for ProtoMetricTagValue_DynamicValue.
var (
	ProtoMetricTagValue_DynamicValue_name = map[int32]string{
		0: "DynamicValueInvalid",
		1: "INDEX",
		2: "INSTANCE_GUID",
	}
	ProtoMetricTagValue_DynamicValue_value = map[string]int32{
		"DynamicValueInvalid": 0,
		"INDEX":               1,
		"INSTANCE_GUID":       2,
	}
)

func (x ProtoMetricTagValue_DynamicValue) Enum() *ProtoMetricTagValue_DynamicValue {
	p := new(ProtoMetricTagValue_DynamicValue)
	*p = x
	return p
}

func (x ProtoMetricTagValue_DynamicValue) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProtoMetricTagValue_DynamicValue) Descriptor() protoreflect.EnumDescriptor {
	return file_metric_tags_proto_enumTypes[0].Descriptor()
}

func (ProtoMetricTagValue_DynamicValue) Type() protoreflect.EnumType {
	return &file_metric_tags_proto_enumTypes[0]
}

func (x ProtoMetricTagValue_DynamicValue) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProtoMetricTagValue_DynamicValue.Descriptor instead.
func (ProtoMetricTagValue_DynamicValue) EnumDescriptor() ([]byte, []int) {
	return file_metric_tags_proto_rawDescGZIP(), []int{0, 0}
}

type ProtoMetricTagValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Note: we only expect one of the following set of fields to be
	// set.
	Static  string                           `protobuf:"bytes,1,opt,name=static,proto3" json:"static,omitempty"`
	Dynamic ProtoMetricTagValue_DynamicValue `protobuf:"varint,2,opt,name=dynamic,proto3,enum=models.ProtoMetricTagValue_DynamicValue" json:"dynamic,omitempty"`
}

func (x *ProtoMetricTagValue) Reset() {
	*x = ProtoMetricTagValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metric_tags_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoMetricTagValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoMetricTagValue) ProtoMessage() {}

func (x *ProtoMetricTagValue) ProtoReflect() protoreflect.Message {
	mi := &file_metric_tags_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoMetricTagValue.ProtoReflect.Descriptor instead.
func (*ProtoMetricTagValue) Descriptor() ([]byte, []int) {
	return file_metric_tags_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoMetricTagValue) GetStatic() string {
	if x != nil {
		return x.Static
	}
	return ""
}

func (x *ProtoMetricTagValue) GetDynamic() ProtoMetricTagValue_DynamicValue {
	if x != nil {
		return x.Dynamic
	}
	return ProtoMetricTagValue_DynamicValueInvalid
}

var File_metric_tags_proto protoreflect.FileDescriptor

var file_metric_tags_proto_rawDesc = []byte{
	0x0a, 0x11, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f, 0x74, 0x61, 0x67, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfe, 0x01, 0x0a, 0x13, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x69, 0x63, 0x12, 0x42, 0x0a, 0x07, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69,
	0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x2e, 0x44, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x52, 0x07, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x22, 0x8a, 0x01, 0x0a, 0x0c, 0x44,
	0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x17, 0x0a, 0x13, 0x44,
	0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x10, 0x00, 0x12, 0x28, 0x0a, 0x05, 0x49, 0x4e, 0x44, 0x45, 0x58, 0x10, 0x01, 0x1a,
	0x1d, 0x82, 0x7d, 0x1a, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x44, 0x79, 0x6e,
	0x61, 0x6d, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x37,
	0x0a, 0x0d, 0x49, 0x4e, 0x53, 0x54, 0x41, 0x4e, 0x43, 0x45, 0x5f, 0x47, 0x55, 0x49, 0x44, 0x10,
	0x02, 0x1a, 0x24, 0x82, 0x7d, 0x21, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x61, 0x67, 0x44,
	0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x47, 0x75, 0x69, 0x64, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67,
	0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_metric_tags_proto_rawDescOnce sync.Once
	file_metric_tags_proto_rawDescData = file_metric_tags_proto_rawDesc
)

func file_metric_tags_proto_rawDescGZIP() []byte {
	file_metric_tags_proto_rawDescOnce.Do(func() {
		file_metric_tags_proto_rawDescData = protoimpl.X.CompressGZIP(file_metric_tags_proto_rawDescData)
	})
	return file_metric_tags_proto_rawDescData
}

var file_metric_tags_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_metric_tags_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_metric_tags_proto_goTypes = []interface{}{
	(ProtoMetricTagValue_DynamicValue)(0), // 0: models.ProtoMetricTagValue.DynamicValue
	(*ProtoMetricTagValue)(nil),           // 1: models.ProtoMetricTagValue
}
var file_metric_tags_proto_depIdxs = []int32{
	0, // 0: models.ProtoMetricTagValue.dynamic:type_name -> models.ProtoMetricTagValue.DynamicValue
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_metric_tags_proto_init() }
func file_metric_tags_proto_init() {
	if File_metric_tags_proto != nil {
		return
	}
	file_bbs_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_metric_tags_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoMetricTagValue); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_metric_tags_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_metric_tags_proto_goTypes,
		DependencyIndexes: file_metric_tags_proto_depIdxs,
		EnumInfos:         file_metric_tags_proto_enumTypes,
		MessageInfos:      file_metric_tags_proto_msgTypes,
	}.Build()
	File_metric_tags_proto = out.File
	file_metric_tags_proto_rawDesc = nil
	file_metric_tags_proto_goTypes = nil
	file_metric_tags_proto_depIdxs = nil
}
