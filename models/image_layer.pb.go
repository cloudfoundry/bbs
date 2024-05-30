// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0
// source: image_layer.proto

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

type ProtoImageLayer_DigestAlgorithm int32

const (
	ProtoImageLayer_DigestAlgorithmInvalid ProtoImageLayer_DigestAlgorithm = 0 // not camel cased since it isn't supposed to be used by API users
	ProtoImageLayer_SHA256                 ProtoImageLayer_DigestAlgorithm = 1
	// Deprecated: Marked as deprecated in image_layer.proto.
	ProtoImageLayer_SHA512 ProtoImageLayer_DigestAlgorithm = 2
)

// Enum value maps for ProtoImageLayer_DigestAlgorithm.
var (
	ProtoImageLayer_DigestAlgorithm_name = map[int32]string{
		0: "DigestAlgorithmInvalid",
		1: "SHA256",
		2: "SHA512",
	}
	ProtoImageLayer_DigestAlgorithm_value = map[string]int32{
		"DigestAlgorithmInvalid": 0,
		"SHA256":                 1,
		"SHA512":                 2,
	}
)

func (x ProtoImageLayer_DigestAlgorithm) Enum() *ProtoImageLayer_DigestAlgorithm {
	p := new(ProtoImageLayer_DigestAlgorithm)
	*p = x
	return p
}

func (x ProtoImageLayer_DigestAlgorithm) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProtoImageLayer_DigestAlgorithm) Descriptor() protoreflect.EnumDescriptor {
	return file_image_layer_proto_enumTypes[0].Descriptor()
}

func (ProtoImageLayer_DigestAlgorithm) Type() protoreflect.EnumType {
	return &file_image_layer_proto_enumTypes[0]
}

func (x ProtoImageLayer_DigestAlgorithm) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProtoImageLayer_DigestAlgorithm.Descriptor instead.
func (ProtoImageLayer_DigestAlgorithm) EnumDescriptor() ([]byte, []int) {
	return file_image_layer_proto_rawDescGZIP(), []int{0, 0}
}

type ProtoImageLayer_MediaType int32

const (
	ProtoImageLayer_MediaTypeInvalid ProtoImageLayer_MediaType = 0 // not camel cased since it isn't supposed to be used by API users
	ProtoImageLayer_TGZ              ProtoImageLayer_MediaType = 1
	ProtoImageLayer_TAR              ProtoImageLayer_MediaType = 2
	ProtoImageLayer_ZIP              ProtoImageLayer_MediaType = 3
)

// Enum value maps for ProtoImageLayer_MediaType.
var (
	ProtoImageLayer_MediaType_name = map[int32]string{
		0: "MediaTypeInvalid",
		1: "TGZ",
		2: "TAR",
		3: "ZIP",
	}
	ProtoImageLayer_MediaType_value = map[string]int32{
		"MediaTypeInvalid": 0,
		"TGZ":              1,
		"TAR":              2,
		"ZIP":              3,
	}
)

func (x ProtoImageLayer_MediaType) Enum() *ProtoImageLayer_MediaType {
	p := new(ProtoImageLayer_MediaType)
	*p = x
	return p
}

func (x ProtoImageLayer_MediaType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProtoImageLayer_MediaType) Descriptor() protoreflect.EnumDescriptor {
	return file_image_layer_proto_enumTypes[1].Descriptor()
}

func (ProtoImageLayer_MediaType) Type() protoreflect.EnumType {
	return &file_image_layer_proto_enumTypes[1]
}

func (x ProtoImageLayer_MediaType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProtoImageLayer_MediaType.Descriptor instead.
func (ProtoImageLayer_MediaType) EnumDescriptor() ([]byte, []int) {
	return file_image_layer_proto_rawDescGZIP(), []int{0, 1}
}

type ProtoImageLayer_Type int32

const (
	ProtoImageLayer_LayerTypeInvalid ProtoImageLayer_Type = 0 // not camel cased since it isn't supposed to be used by API users
	ProtoImageLayer_SHARED           ProtoImageLayer_Type = 1
	ProtoImageLayer_EXCLUSIVE        ProtoImageLayer_Type = 2
)

// Enum value maps for ProtoImageLayer_Type.
var (
	ProtoImageLayer_Type_name = map[int32]string{
		0: "LayerTypeInvalid",
		1: "SHARED",
		2: "EXCLUSIVE",
	}
	ProtoImageLayer_Type_value = map[string]int32{
		"LayerTypeInvalid": 0,
		"SHARED":           1,
		"EXCLUSIVE":        2,
	}
)

func (x ProtoImageLayer_Type) Enum() *ProtoImageLayer_Type {
	p := new(ProtoImageLayer_Type)
	*p = x
	return p
}

func (x ProtoImageLayer_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProtoImageLayer_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_image_layer_proto_enumTypes[2].Descriptor()
}

func (ProtoImageLayer_Type) Type() protoreflect.EnumType {
	return &file_image_layer_proto_enumTypes[2]
}

func (x ProtoImageLayer_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProtoImageLayer_Type.Descriptor instead.
func (ProtoImageLayer_Type) EnumDescriptor() ([]byte, []int) {
	return file_image_layer_proto_rawDescGZIP(), []int{0, 2}
}

type ProtoImageLayer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name            string                          `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Url             string                          `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	DestinationPath string                          `protobuf:"bytes,3,opt,name=destination_path,proto3" json:"destination_path,omitempty"`
	LayerType       ProtoImageLayer_Type            `protobuf:"varint,4,opt,name=layer_type,proto3,enum=models.ProtoImageLayer_Type" json:"layer_type,omitempty"`
	MediaType       ProtoImageLayer_MediaType       `protobuf:"varint,5,opt,name=media_type,proto3,enum=models.ProtoImageLayer_MediaType" json:"media_type,omitempty"`
	DigestAlgorithm ProtoImageLayer_DigestAlgorithm `protobuf:"varint,6,opt,name=digest_algorithm,proto3,enum=models.ProtoImageLayer_DigestAlgorithm" json:"digest_algorithm,omitempty"`
	DigestValue     string                          `protobuf:"bytes,7,opt,name=digest_value,proto3" json:"digest_value,omitempty"`
}

func (x *ProtoImageLayer) Reset() {
	*x = ProtoImageLayer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_image_layer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoImageLayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoImageLayer) ProtoMessage() {}

func (x *ProtoImageLayer) ProtoReflect() protoreflect.Message {
	mi := &file_image_layer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoImageLayer.ProtoReflect.Descriptor instead.
func (*ProtoImageLayer) Descriptor() ([]byte, []int) {
	return file_image_layer_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoImageLayer) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ProtoImageLayer) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *ProtoImageLayer) GetDestinationPath() string {
	if x != nil {
		return x.DestinationPath
	}
	return ""
}

func (x *ProtoImageLayer) GetLayerType() ProtoImageLayer_Type {
	if x != nil {
		return x.LayerType
	}
	return ProtoImageLayer_LayerTypeInvalid
}

func (x *ProtoImageLayer) GetMediaType() ProtoImageLayer_MediaType {
	if x != nil {
		return x.MediaType
	}
	return ProtoImageLayer_MediaTypeInvalid
}

func (x *ProtoImageLayer) GetDigestAlgorithm() ProtoImageLayer_DigestAlgorithm {
	if x != nil {
		return x.DigestAlgorithm
	}
	return ProtoImageLayer_DigestAlgorithmInvalid
}

func (x *ProtoImageLayer) GetDigestValue() string {
	if x != nil {
		return x.DigestValue
	}
	return ""
}

var File_image_layer_proto protoreflect.FileDescriptor

var file_image_layer_proto_rawDesc = []byte{
	0x0a, 0x11, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc3, 0x05, 0x0a, 0x0f, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x49, 0x6d, 0x61, 0x67, 0x65, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x15,
	0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01,
	0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x2f, 0x0a, 0x10, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x03, 0xc0, 0x3e, 0x01, 0x52, 0x10, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x12, 0x41, 0x0a, 0x0a, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4c, 0x61,
	0x79, 0x65, 0x72, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0a, 0x6c,
	0x61, 0x79, 0x65, 0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x12, 0x46, 0x0a, 0x0a, 0x6d, 0x65, 0x64,
	0x69, 0x61, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x49, 0x6d, 0x61, 0x67,
	0x65, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79, 0x70, 0x65,
	0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0a, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x53, 0x0a, 0x10, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x5f, 0x61, 0x6c, 0x67, 0x6f,
	0x72, 0x69, 0x74, 0x68, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4c,
	0x61, 0x79, 0x65, 0x72, 0x2e, 0x44, 0x69, 0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72,
	0x69, 0x74, 0x68, 0x6d, 0x52, 0x10, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x5f, 0x61, 0x6c, 0x67,
	0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x22, 0x0a, 0x0c, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x64, 0x69,
	0x67, 0x65, 0x73, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x7b, 0x0a, 0x0f, 0x44, 0x69,
	0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x1a, 0x0a,
	0x16, 0x44, 0x69, 0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d,
	0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x10, 0x00, 0x12, 0x24, 0x0a, 0x06, 0x53, 0x48, 0x41,
	0x32, 0x35, 0x36, 0x10, 0x01, 0x1a, 0x18, 0x82, 0x7d, 0x15, 0x44, 0x69, 0x67, 0x65, 0x73, 0x74,
	0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x53, 0x68, 0x61, 0x32, 0x35, 0x36, 0x12,
	0x26, 0x0a, 0x06, 0x53, 0x48, 0x41, 0x35, 0x31, 0x32, 0x10, 0x02, 0x1a, 0x1a, 0x82, 0x7d, 0x15,
	0x44, 0x69, 0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x53,
	0x68, 0x61, 0x35, 0x31, 0x32, 0x08, 0x01, 0x22, 0x6f, 0x0a, 0x09, 0x4d, 0x65, 0x64, 0x69, 0x61,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79, 0x70,
	0x65, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x10, 0x00, 0x12, 0x18, 0x0a, 0x03, 0x54, 0x47,
	0x5a, 0x10, 0x01, 0x1a, 0x0f, 0x82, 0x7d, 0x0c, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79, 0x70,
	0x65, 0x54, 0x67, 0x7a, 0x12, 0x18, 0x0a, 0x03, 0x54, 0x41, 0x52, 0x10, 0x02, 0x1a, 0x0f, 0x82,
	0x7d, 0x0c, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x54, 0x79, 0x70, 0x65, 0x54, 0x61, 0x72, 0x12, 0x18,
	0x0a, 0x03, 0x5a, 0x49, 0x50, 0x10, 0x03, 0x1a, 0x0f, 0x82, 0x7d, 0x0c, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x54, 0x79, 0x70, 0x65, 0x5a, 0x69, 0x70, 0x22, 0x62, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x14, 0x0a, 0x10, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x49, 0x6e, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x10, 0x00, 0x12, 0x1e, 0x0a, 0x06, 0x53, 0x48, 0x41, 0x52, 0x45, 0x44,
	0x10, 0x01, 0x1a, 0x12, 0x82, 0x7d, 0x0f, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65,
	0x53, 0x68, 0x61, 0x72, 0x65, 0x64, 0x12, 0x24, 0x0a, 0x09, 0x45, 0x58, 0x43, 0x4c, 0x55, 0x53,
	0x49, 0x56, 0x45, 0x10, 0x02, 0x1a, 0x15, 0x82, 0x7d, 0x12, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x45, 0x78, 0x63, 0x6c, 0x75, 0x73, 0x69, 0x76, 0x65, 0x42, 0x22, 0x5a, 0x20,
	0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72,
	0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_image_layer_proto_rawDescOnce sync.Once
	file_image_layer_proto_rawDescData = file_image_layer_proto_rawDesc
)

func file_image_layer_proto_rawDescGZIP() []byte {
	file_image_layer_proto_rawDescOnce.Do(func() {
		file_image_layer_proto_rawDescData = protoimpl.X.CompressGZIP(file_image_layer_proto_rawDescData)
	})
	return file_image_layer_proto_rawDescData
}

var file_image_layer_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_image_layer_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_image_layer_proto_goTypes = []interface{}{
	(ProtoImageLayer_DigestAlgorithm)(0), // 0: models.ProtoImageLayer.DigestAlgorithm
	(ProtoImageLayer_MediaType)(0),       // 1: models.ProtoImageLayer.MediaType
	(ProtoImageLayer_Type)(0),            // 2: models.ProtoImageLayer.Type
	(*ProtoImageLayer)(nil),              // 3: models.ProtoImageLayer
}
var file_image_layer_proto_depIdxs = []int32{
	2, // 0: models.ProtoImageLayer.layer_type:type_name -> models.ProtoImageLayer.Type
	1, // 1: models.ProtoImageLayer.media_type:type_name -> models.ProtoImageLayer.MediaType
	0, // 2: models.ProtoImageLayer.digest_algorithm:type_name -> models.ProtoImageLayer.DigestAlgorithm
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_image_layer_proto_init() }
func file_image_layer_proto_init() {
	if File_image_layer_proto != nil {
		return
	}
	file_bbs_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_image_layer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoImageLayer); i {
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
			RawDescriptor: file_image_layer_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_image_layer_proto_goTypes,
		DependencyIndexes: file_image_layer_proto_depIdxs,
		EnumInfos:         file_image_layer_proto_enumTypes,
		MessageInfos:      file_image_layer_proto_msgTypes,
	}.Build()
	File_image_layer_proto = out.File
	file_image_layer_proto_rawDesc = nil
	file_image_layer_proto_goTypes = nil
	file_image_layer_proto_depIdxs = nil
}
