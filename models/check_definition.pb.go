// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: check_definition.proto

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

type CheckDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Checks          []*Check `protobuf:"bytes,1,rep,name=checks,proto3" json:"checks,omitempty"`
	LogSource       string   `protobuf:"bytes,2,opt,name=log_source,proto3" json:"log_source,omitempty"`
	ReadinessChecks []*Check `protobuf:"bytes,3,rep,name=readiness_checks,json=readinessChecks,proto3" json:"readiness_checks,omitempty"`
}

func (x *CheckDefinition) Reset() {
	*x = CheckDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_check_definition_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckDefinition) ProtoMessage() {}

func (x *CheckDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_check_definition_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckDefinition.ProtoReflect.Descriptor instead.
func (*CheckDefinition) Descriptor() ([]byte, []int) {
	return file_check_definition_proto_rawDescGZIP(), []int{0}
}

func (x *CheckDefinition) GetChecks() []*Check {
	if x != nil {
		return x.Checks
	}
	return nil
}

func (x *CheckDefinition) GetLogSource() string {
	if x != nil {
		return x.LogSource
	}
	return ""
}

func (x *CheckDefinition) GetReadinessChecks() []*Check {
	if x != nil {
		return x.ReadinessChecks
	}
	return nil
}

type Check struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// oneof is hard to use right now, instead we can do this check in validation
	// oneof check {
	TcpCheck  *TCPCheck  `protobuf:"bytes,1,opt,name=tcp_check,json=tcpCheck,proto3" json:"tcp_check,omitempty"`
	HttpCheck *HTTPCheck `protobuf:"bytes,2,opt,name=http_check,json=httpCheck,proto3" json:"http_check,omitempty"` // }
}

func (x *Check) Reset() {
	*x = Check{}
	if protoimpl.UnsafeEnabled {
		mi := &file_check_definition_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Check) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Check) ProtoMessage() {}

func (x *Check) ProtoReflect() protoreflect.Message {
	mi := &file_check_definition_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Check.ProtoReflect.Descriptor instead.
func (*Check) Descriptor() ([]byte, []int) {
	return file_check_definition_proto_rawDescGZIP(), []int{1}
}

func (x *Check) GetTcpCheck() *TCPCheck {
	if x != nil {
		return x.TcpCheck
	}
	return nil
}

func (x *Check) GetHttpCheck() *HTTPCheck {
	if x != nil {
		return x.HttpCheck
	}
	return nil
}

type TCPCheck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Port             uint32 `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	ConnectTimeoutMs uint64 `protobuf:"varint,2,opt,name=connect_timeout_ms,json=connectTimeoutMs,proto3" json:"connect_timeout_ms,omitempty"`
	IntervalMs       uint64 `protobuf:"varint,3,opt,name=interval_ms,json=intervalMs,proto3" json:"interval_ms,omitempty"`
}

func (x *TCPCheck) Reset() {
	*x = TCPCheck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_check_definition_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TCPCheck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TCPCheck) ProtoMessage() {}

func (x *TCPCheck) ProtoReflect() protoreflect.Message {
	mi := &file_check_definition_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TCPCheck.ProtoReflect.Descriptor instead.
func (*TCPCheck) Descriptor() ([]byte, []int) {
	return file_check_definition_proto_rawDescGZIP(), []int{2}
}

func (x *TCPCheck) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *TCPCheck) GetConnectTimeoutMs() uint64 {
	if x != nil {
		return x.ConnectTimeoutMs
	}
	return 0
}

func (x *TCPCheck) GetIntervalMs() uint64 {
	if x != nil {
		return x.IntervalMs
	}
	return 0
}

type HTTPCheck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Port             uint32 `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	RequestTimeoutMs uint64 `protobuf:"varint,2,opt,name=request_timeout_ms,json=requestTimeoutMs,proto3" json:"request_timeout_ms,omitempty"`
	Path             string `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	IntervalMs       uint64 `protobuf:"varint,4,opt,name=interval_ms,json=intervalMs,proto3" json:"interval_ms,omitempty"`
}

func (x *HTTPCheck) Reset() {
	*x = HTTPCheck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_check_definition_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HTTPCheck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HTTPCheck) ProtoMessage() {}

func (x *HTTPCheck) ProtoReflect() protoreflect.Message {
	mi := &file_check_definition_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HTTPCheck.ProtoReflect.Descriptor instead.
func (*HTTPCheck) Descriptor() ([]byte, []int) {
	return file_check_definition_proto_rawDescGZIP(), []int{3}
}

func (x *HTTPCheck) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *HTTPCheck) GetRequestTimeoutMs() uint64 {
	if x != nil {
		return x.RequestTimeoutMs
	}
	return 0
}

func (x *HTTPCheck) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *HTTPCheck) GetIntervalMs() uint64 {
	if x != nil {
		return x.IntervalMs
	}
	return 0
}

var File_check_definition_proto protoreflect.FileDescriptor

var file_check_definition_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x22, 0x92, 0x01, 0x0a, 0x0f, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x06, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x06, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x6c,
	0x6f, 0x67, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x6c, 0x6f, 0x67, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x38, 0x0a, 0x10, 0x72,
	0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x52, 0x0f, 0x72, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x73, 0x22, 0x68, 0x0a, 0x05, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x2d,
	0x0a, 0x09, 0x74, 0x63, 0x70, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x54, 0x43, 0x50, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x08, 0x74, 0x63, 0x70, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x30, 0x0a,
	0x0a, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x48, 0x54, 0x54, 0x50, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x52, 0x09, 0x68, 0x74, 0x74, 0x70, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x22,
	0x6d, 0x0a, 0x08, 0x54, 0x43, 0x50, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x6f, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12,
	0x2c, 0x0a, 0x12, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x10, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x12, 0x1f, 0x0a,
	0x0b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x5f, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x4d, 0x73, 0x22, 0x82,
	0x01, 0x0a, 0x09, 0x48, 0x54, 0x54, 0x50, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74,
	0x12, 0x2c, 0x0a, 0x12, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x6f, 0x75, 0x74, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x10, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61,
	0x74, 0x68, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x5f, 0x6d,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61,
	0x6c, 0x4d, 0x73, 0x42, 0x22, 0x5a, 0x20, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x72, 0x79, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x62, 0x62, 0x73,
	0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_check_definition_proto_rawDescOnce sync.Once
	file_check_definition_proto_rawDescData = file_check_definition_proto_rawDesc
)

func file_check_definition_proto_rawDescGZIP() []byte {
	file_check_definition_proto_rawDescOnce.Do(func() {
		file_check_definition_proto_rawDescData = protoimpl.X.CompressGZIP(file_check_definition_proto_rawDescData)
	})
	return file_check_definition_proto_rawDescData
}

var file_check_definition_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_check_definition_proto_goTypes = []interface{}{
	(*CheckDefinition)(nil), // 0: models.CheckDefinition
	(*Check)(nil),           // 1: models.Check
	(*TCPCheck)(nil),        // 2: models.TCPCheck
	(*HTTPCheck)(nil),       // 3: models.HTTPCheck
}
var file_check_definition_proto_depIdxs = []int32{
	1, // 0: models.CheckDefinition.checks:type_name -> models.Check
	1, // 1: models.CheckDefinition.readiness_checks:type_name -> models.Check
	2, // 2: models.Check.tcp_check:type_name -> models.TCPCheck
	3, // 3: models.Check.http_check:type_name -> models.HTTPCheck
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_check_definition_proto_init() }
func file_check_definition_proto_init() {
	if File_check_definition_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_check_definition_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckDefinition); i {
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
		file_check_definition_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Check); i {
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
		file_check_definition_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TCPCheck); i {
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
		file_check_definition_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HTTPCheck); i {
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
			RawDescriptor: file_check_definition_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_check_definition_proto_goTypes,
		DependencyIndexes: file_check_definition_proto_depIdxs,
		MessageInfos:      file_check_definition_proto_msgTypes,
	}.Build()
	File_check_definition_proto = out.File
	file_check_definition_proto_rawDesc = nil
	file_check_definition_proto_goTypes = nil
	file_check_definition_proto_depIdxs = nil
}
