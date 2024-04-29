// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.0--rc1
// source: cells.proto

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

type ProtoCellCapacity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MemoryMb   int32 `protobuf:"varint,1,opt,name=memory_mb,proto3" json:"memory_mb,omitempty"`
	DiskMb     int32 `protobuf:"varint,2,opt,name=disk_mb,proto3" json:"disk_mb,omitempty"`
	Containers int32 `protobuf:"varint,3,opt,name=containers,proto3" json:"containers,omitempty"`
}

func (x *ProtoCellCapacity) Reset() {
	*x = ProtoCellCapacity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoCellCapacity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoCellCapacity) ProtoMessage() {}

func (x *ProtoCellCapacity) ProtoReflect() protoreflect.Message {
	mi := &file_cells_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoCellCapacity.ProtoReflect.Descriptor instead.
func (*ProtoCellCapacity) Descriptor() ([]byte, []int) {
	return file_cells_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoCellCapacity) GetMemoryMb() int32 {
	if x != nil {
		return x.MemoryMb
	}
	return 0
}

func (x *ProtoCellCapacity) GetDiskMb() int32 {
	if x != nil {
		return x.DiskMb
	}
	return 0
}

func (x *ProtoCellCapacity) GetContainers() int32 {
	if x != nil {
		return x.Containers
	}
	return 0
}

type ProtoCellPresence struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CellId                string             `protobuf:"bytes,1,opt,name=cell_id,proto3" json:"cell_id,omitempty"`
	RepAddress            string             `protobuf:"bytes,2,opt,name=rep_address,proto3" json:"rep_address,omitempty"`
	Zone                  string             `protobuf:"bytes,3,opt,name=zone,proto3" json:"zone,omitempty"`
	Capacity              *ProtoCellCapacity `protobuf:"bytes,4,opt,name=capacity,proto3" json:"capacity,omitempty"`
	RootfsProviders       []*ProtoProvider   `protobuf:"bytes,5,rep,name=rootfs_providers,json=rootfs_provider_list,proto3" json:"rootfs_providers,omitempty"`
	PlacementTags         []string           `protobuf:"bytes,6,rep,name=placement_tags,json=placementTags,proto3" json:"placement_tags,omitempty"`
	OptionalPlacementTags []string           `protobuf:"bytes,7,rep,name=optional_placement_tags,json=optionalPlacementTags,proto3" json:"optional_placement_tags,omitempty"`
	RepUrl                string             `protobuf:"bytes,8,opt,name=rep_url,proto3" json:"rep_url,omitempty"`
}

func (x *ProtoCellPresence) Reset() {
	*x = ProtoCellPresence{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoCellPresence) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoCellPresence) ProtoMessage() {}

func (x *ProtoCellPresence) ProtoReflect() protoreflect.Message {
	mi := &file_cells_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoCellPresence.ProtoReflect.Descriptor instead.
func (*ProtoCellPresence) Descriptor() ([]byte, []int) {
	return file_cells_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoCellPresence) GetCellId() string {
	if x != nil {
		return x.CellId
	}
	return ""
}

func (x *ProtoCellPresence) GetRepAddress() string {
	if x != nil {
		return x.RepAddress
	}
	return ""
}

func (x *ProtoCellPresence) GetZone() string {
	if x != nil {
		return x.Zone
	}
	return ""
}

func (x *ProtoCellPresence) GetCapacity() *ProtoCellCapacity {
	if x != nil {
		return x.Capacity
	}
	return nil
}

func (x *ProtoCellPresence) GetRootfsProviders() []*ProtoProvider {
	if x != nil {
		return x.RootfsProviders
	}
	return nil
}

func (x *ProtoCellPresence) GetPlacementTags() []string {
	if x != nil {
		return x.PlacementTags
	}
	return nil
}

func (x *ProtoCellPresence) GetOptionalPlacementTags() []string {
	if x != nil {
		return x.OptionalPlacementTags
	}
	return nil
}

func (x *ProtoCellPresence) GetRepUrl() string {
	if x != nil {
		return x.RepUrl
	}
	return ""
}

type ProtoProvider struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Properties []string `protobuf:"bytes,2,rep,name=properties,proto3" json:"properties,omitempty"`
}

func (x *ProtoProvider) Reset() {
	*x = ProtoProvider{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoProvider) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoProvider) ProtoMessage() {}

func (x *ProtoProvider) ProtoReflect() protoreflect.Message {
	mi := &file_cells_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoProvider.ProtoReflect.Descriptor instead.
func (*ProtoProvider) Descriptor() ([]byte, []int) {
	return file_cells_proto_rawDescGZIP(), []int{2}
}

func (x *ProtoProvider) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ProtoProvider) GetProperties() []string {
	if x != nil {
		return x.Properties
	}
	return nil
}

type ProtoCellsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *ProtoError          `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Cells []*ProtoCellPresence `protobuf:"bytes,2,rep,name=cells,proto3" json:"cells,omitempty"`
}

func (x *ProtoCellsResponse) Reset() {
	*x = ProtoCellsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cells_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoCellsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoCellsResponse) ProtoMessage() {}

func (x *ProtoCellsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cells_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoCellsResponse.ProtoReflect.Descriptor instead.
func (*ProtoCellsResponse) Descriptor() ([]byte, []int) {
	return file_cells_proto_rawDescGZIP(), []int{3}
}

func (x *ProtoCellsResponse) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoCellsResponse) GetCells() []*ProtoCellPresence {
	if x != nil {
		return x.Cells
	}
	return nil
}

var File_cells_proto protoreflect.FileDescriptor

var file_cells_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x6b, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43, 0x65, 0x6c, 0x6c, 0x43,
	0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x65, 0x6d, 0x6f, 0x72,
	0x79, 0x5f, 0x6d, 0x62, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x6d, 0x65, 0x6d, 0x6f,
	0x72, 0x79, 0x5f, 0x6d, 0x62, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x6d, 0x62,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x6d, 0x62, 0x12,
	0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x22,
	0xda, 0x02, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x65,
	0x73, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x12,
	0x20, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x12, 0x0a, 0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x7a, 0x6f, 0x6e, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43, 0x65, 0x6c, 0x6c, 0x43, 0x61, 0x70, 0x61, 0x63, 0x69,
	0x74, 0x79, 0x52, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x45, 0x0a, 0x10,
	0x72, 0x6f, 0x6f, 0x74, 0x66, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x14, 0x72,
	0x6f, 0x6f, 0x74, 0x66, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x6c,
	0x69, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x6c, 0x61,
	0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x54, 0x61, 0x67, 0x73, 0x12, 0x36, 0x0a, 0x17, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x5f, 0x74, 0x61, 0x67, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x15, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x50, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x54, 0x61,
	0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x70, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x70, 0x5f, 0x75, 0x72, 0x6c, 0x22, 0x43, 0x0a, 0x0d,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65,
	0x73, 0x22, 0x6f, 0x0a, 0x12, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43, 0x65, 0x6c, 0x6c, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x12, 0x2f, 0x0a, 0x05, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x43,
	0x65, 0x6c, 0x6c, 0x50, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x05, 0x63, 0x65, 0x6c,
	0x6c, 0x73, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cells_proto_rawDescOnce sync.Once
	file_cells_proto_rawDescData = file_cells_proto_rawDesc
)

func file_cells_proto_rawDescGZIP() []byte {
	file_cells_proto_rawDescOnce.Do(func() {
		file_cells_proto_rawDescData = protoimpl.X.CompressGZIP(file_cells_proto_rawDescData)
	})
	return file_cells_proto_rawDescData
}

var file_cells_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_cells_proto_goTypes = []interface{}{
	(*ProtoCellCapacity)(nil),  // 0: models.ProtoCellCapacity
	(*ProtoCellPresence)(nil),  // 1: models.ProtoCellPresence
	(*ProtoProvider)(nil),      // 2: models.ProtoProvider
	(*ProtoCellsResponse)(nil), // 3: models.ProtoCellsResponse
	(*ProtoError)(nil),         // 4: models.ProtoError
}
var file_cells_proto_depIdxs = []int32{
	0, // 0: models.ProtoCellPresence.capacity:type_name -> models.ProtoCellCapacity
	2, // 1: models.ProtoCellPresence.rootfs_providers:type_name -> models.ProtoProvider
	4, // 2: models.ProtoCellsResponse.error:type_name -> models.ProtoError
	1, // 3: models.ProtoCellsResponse.cells:type_name -> models.ProtoCellPresence
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_cells_proto_init() }
func file_cells_proto_init() {
	if File_cells_proto != nil {
		return
	}
	file_error_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_cells_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoCellCapacity); i {
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
		file_cells_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoCellPresence); i {
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
		file_cells_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoProvider); i {
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
		file_cells_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoCellsResponse); i {
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
			RawDescriptor: file_cells_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cells_proto_goTypes,
		DependencyIndexes: file_cells_proto_depIdxs,
		MessageInfos:      file_cells_proto_msgTypes,
	}.Build()
	File_cells_proto = out.File
	file_cells_proto_rawDesc = nil
	file_cells_proto_goTypes = nil
	file_cells_proto_depIdxs = nil
}
