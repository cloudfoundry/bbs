// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.2
// source: volume_mount.proto

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

type ProtoSharedDevice struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VolumeId    string `protobuf:"bytes,1,opt,name=volume_id,proto3" json:"volume_id,omitempty"`
	MountConfig string `protobuf:"bytes,2,opt,name=mount_config,proto3" json:"mount_config,omitempty"`
}

func (x *ProtoSharedDevice) Reset() {
	*x = ProtoSharedDevice{}
	mi := &file_volume_mount_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoSharedDevice) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoSharedDevice) ProtoMessage() {}

func (x *ProtoSharedDevice) ProtoReflect() protoreflect.Message {
	mi := &file_volume_mount_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoSharedDevice.ProtoReflect.Descriptor instead.
func (*ProtoSharedDevice) Descriptor() ([]byte, []int) {
	return file_volume_mount_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoSharedDevice) GetVolumeId() string {
	if x != nil {
		return x.VolumeId
	}
	return ""
}

func (x *ProtoSharedDevice) GetMountConfig() string {
	if x != nil {
		return x.MountConfig
	}
	return ""
}

type ProtoVolumeMount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Driver       string `protobuf:"bytes,1,opt,name=driver,proto3" json:"driver,omitempty"`
	ContainerDir string `protobuf:"bytes,3,opt,name=container_dir,proto3" json:"container_dir,omitempty"`
	Mode         string `protobuf:"bytes,6,opt,name=mode,proto3" json:"mode,omitempty"`
	// oneof device {
	Shared *ProtoSharedDevice `protobuf:"bytes,7,opt,name=shared,proto3" json:"shared,omitempty"` // }
}

func (x *ProtoVolumeMount) Reset() {
	*x = ProtoVolumeMount{}
	mi := &file_volume_mount_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoVolumeMount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoVolumeMount) ProtoMessage() {}

func (x *ProtoVolumeMount) ProtoReflect() protoreflect.Message {
	mi := &file_volume_mount_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoVolumeMount.ProtoReflect.Descriptor instead.
func (*ProtoVolumeMount) Descriptor() ([]byte, []int) {
	return file_volume_mount_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoVolumeMount) GetDriver() string {
	if x != nil {
		return x.Driver
	}
	return ""
}

func (x *ProtoVolumeMount) GetContainerDir() string {
	if x != nil {
		return x.ContainerDir
	}
	return ""
}

func (x *ProtoVolumeMount) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

func (x *ProtoVolumeMount) GetShared() *ProtoSharedDevice {
	if x != nil {
		return x.Shared
	}
	return nil
}

type ProtoVolumePlacement struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DriverNames []string `protobuf:"bytes,1,rep,name=driver_names,proto3" json:"driver_names,omitempty"`
}

func (x *ProtoVolumePlacement) Reset() {
	*x = ProtoVolumePlacement{}
	mi := &file_volume_mount_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProtoVolumePlacement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoVolumePlacement) ProtoMessage() {}

func (x *ProtoVolumePlacement) ProtoReflect() protoreflect.Message {
	mi := &file_volume_mount_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoVolumePlacement.ProtoReflect.Descriptor instead.
func (*ProtoVolumePlacement) Descriptor() ([]byte, []int) {
	return file_volume_mount_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoVolumePlacement) GetDriverNames() []string {
	if x != nil {
		return x.DriverNames
	}
	return nil
}

var File_volume_mount_proto protoreflect.FileDescriptor

var file_volume_mount_proto_rawDesc = []byte{
	0x0a, 0x12, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x5f, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x09, 0x62, 0x62,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5f, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x53, 0x68, 0x61, 0x72, 0x65, 0x64, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x21, 0x0a, 0x09,
	0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x03, 0xc0, 0x3e, 0x01, 0x52, 0x09, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x12,
	0x27, 0x0a, 0x0c, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x6d, 0x6f, 0x75, 0x6e,
	0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xbd, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x4d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1b, 0x0a,
	0x06, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xc0,
	0x3e, 0x01, 0x52, 0x06, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x0d, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x64, 0x69, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65,
	0x72, 0x5f, 0x64, 0x69, 0x72, 0x12, 0x17, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x36,
	0x0a, 0x06, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x53, 0x68, 0x61,
	0x72, 0x65, 0x64, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x06,
	0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x4a, 0x04, 0x08, 0x04,
	0x10, 0x05, 0x4a, 0x04, 0x08, 0x05, 0x10, 0x06, 0x22, 0x3f, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x56, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x12, 0x27, 0x0a, 0x0c, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x42, 0x03, 0xc0, 0x3e, 0x01, 0x52, 0x0c, 0x64, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64,
	0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f,
	0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_volume_mount_proto_rawDescOnce sync.Once
	file_volume_mount_proto_rawDescData = file_volume_mount_proto_rawDesc
)

func file_volume_mount_proto_rawDescGZIP() []byte {
	file_volume_mount_proto_rawDescOnce.Do(func() {
		file_volume_mount_proto_rawDescData = protoimpl.X.CompressGZIP(file_volume_mount_proto_rawDescData)
	})
	return file_volume_mount_proto_rawDescData
}

var file_volume_mount_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_volume_mount_proto_goTypes = []any{
	(*ProtoSharedDevice)(nil),    // 0: models.ProtoSharedDevice
	(*ProtoVolumeMount)(nil),     // 1: models.ProtoVolumeMount
	(*ProtoVolumePlacement)(nil), // 2: models.ProtoVolumePlacement
}
var file_volume_mount_proto_depIdxs = []int32{
	0, // 0: models.ProtoVolumeMount.shared:type_name -> models.ProtoSharedDevice
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_volume_mount_proto_init() }
func file_volume_mount_proto_init() {
	if File_volume_mount_proto != nil {
		return
	}
	file_bbs_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_volume_mount_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_volume_mount_proto_goTypes,
		DependencyIndexes: file_volume_mount_proto_depIdxs,
		MessageInfos:      file_volume_mount_proto_msgTypes,
	}.Build()
	File_volume_mount_proto = out.File
	file_volume_mount_proto_rawDesc = nil
	file_volume_mount_proto_goTypes = nil
	file_volume_mount_proto_depIdxs = nil
}
