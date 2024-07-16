// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.0
// source: bbs.proto

package main

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_bbs_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         1000,
		Name:          "bbs.bbs_json_always_emit",
		Tag:           "varint,1000,opt,name=bbs_json_always_emit",
		Filename:      "bbs.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         1010,
		Name:          "bbs.bbs_by_value",
		Tag:           "varint,1010,opt,name=bbs_by_value",
		Filename:      "bbs.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         1020,
		Name:          "bbs.bbs_exclude_from_equal",
		Tag:           "varint,1020,opt,name=bbs_exclude_from_equal",
		Filename:      "bbs.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         1030,
		Name:          "bbs.bbs_custom_type",
		Tag:           "bytes,1030,opt,name=bbs_custom_type",
		Filename:      "bbs.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         2000,
		Name:          "bbs.bbs_enumvalue_customname",
		Tag:           "bytes,2000,opt,name=bbs_enumvalue_customname",
		Filename:      "bbs.proto",
	},
}

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional bool bbs_json_always_emit = 1000;
	E_BbsJsonAlwaysEmit = &file_bbs_proto_extTypes[0]
	// optional bool bbs_by_value = 1010;
	E_BbsByValue = &file_bbs_proto_extTypes[1]
	// optional bool bbs_exclude_from_equal = 1020;
	E_BbsExcludeFromEqual = &file_bbs_proto_extTypes[2]
	// optional string bbs_custom_type = 1030;
	E_BbsCustomType = &file_bbs_proto_extTypes[3]
)

// Extension fields to descriptorpb.EnumValueOptions.
var (
	// optional string bbs_enumvalue_customname = 2000;
	E_BbsEnumvalueCustomname = &file_bbs_proto_extTypes[4]
)

var File_bbs_proto protoreflect.FileDescriptor

var file_bbs_proto_rawDesc = []byte{
	0x0a, 0x09, 0x62, 0x62, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x62, 0x62, 0x73,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x3a, 0x52, 0x0a, 0x14, 0x62, 0x62, 0x73, 0x5f, 0x6a, 0x73, 0x6f, 0x6e, 0x5f, 0x61,
	0x6c, 0x77, 0x61, 0x79, 0x73, 0x5f, 0x65, 0x6d, 0x69, 0x74, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe8, 0x07, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x11, 0x62, 0x62, 0x73, 0x4a, 0x73, 0x6f, 0x6e, 0x41, 0x6c, 0x77, 0x61, 0x79, 0x73, 0x45,
	0x6d, 0x69, 0x74, 0x88, 0x01, 0x01, 0x3a, 0x43, 0x0a, 0x0c, 0x62, 0x62, 0x73, 0x5f, 0x62, 0x79,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf2, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x62, 0x62,
	0x73, 0x42, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x88, 0x01, 0x01, 0x3a, 0x56, 0x0a, 0x16, 0x62,
	0x62, 0x73, 0x5f, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x5f,
	0x65, 0x71, 0x75, 0x61, 0x6c, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xfc, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13, 0x62, 0x62, 0x73,
	0x45, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x45, 0x71, 0x75, 0x61, 0x6c,
	0x88, 0x01, 0x01, 0x3a, 0x49, 0x0a, 0x0f, 0x62, 0x62, 0x73, 0x5f, 0x63, 0x75, 0x73, 0x74, 0x6f,
	0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x86, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x62, 0x62,
	0x73, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x3a, 0x5f,
	0x0a, 0x18, 0x62, 0x62, 0x73, 0x5f, 0x65, 0x6e, 0x75, 0x6d, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6e, 0x75,
	0x6d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x0f,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x62, 0x62, 0x73, 0x45, 0x6e, 0x75, 0x6d, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x42,
	0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75,
	0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_bbs_proto_goTypes = []any{
	(*descriptorpb.FieldOptions)(nil),     // 0: google.protobuf.FieldOptions
	(*descriptorpb.EnumValueOptions)(nil), // 1: google.protobuf.EnumValueOptions
}
var file_bbs_proto_depIdxs = []int32{
	0, // 0: bbs.bbs_json_always_emit:extendee -> google.protobuf.FieldOptions
	0, // 1: bbs.bbs_by_value:extendee -> google.protobuf.FieldOptions
	0, // 2: bbs.bbs_exclude_from_equal:extendee -> google.protobuf.FieldOptions
	0, // 3: bbs.bbs_custom_type:extendee -> google.protobuf.FieldOptions
	1, // 4: bbs.bbs_enumvalue_customname:extendee -> google.protobuf.EnumValueOptions
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	0, // [0:5] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_bbs_proto_init() }
func file_bbs_proto_init() {
	if File_bbs_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_bbs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 5,
			NumServices:   0,
		},
		GoTypes:           file_bbs_proto_goTypes,
		DependencyIndexes: file_bbs_proto_depIdxs,
		ExtensionInfos:    file_bbs_proto_extTypes,
	}.Build()
	File_bbs_proto = out.File
	file_bbs_proto_rawDesc = nil
	file_bbs_proto_goTypes = nil
	file_bbs_proto_depIdxs = nil
}
