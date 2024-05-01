// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0--rc1
// source: error.proto

package models

type Error_Type int32

const (
	Error_UnknownError               Error_Type = 0
	Error_InvalidRecord              Error_Type = 3
	Error_InvalidRequest             Error_Type = 4
	Error_InvalidResponse            Error_Type = 5
	Error_InvalidProtobufMessage     Error_Type = 6
	Error_InvalidJSON                Error_Type = 7
	Error_FailedToOpenEnvelope       Error_Type = 8
	Error_InvalidStateTransition     Error_Type = 9
	Error_ResourceConflict           Error_Type = 11
	Error_ResourceExists             Error_Type = 12
	Error_ResourceNotFound           Error_Type = 13
	Error_RouterError                Error_Type = 14
	Error_ActualLRPCannotBeClaimed   Error_Type = 15
	Error_ActualLRPCannotBeStarted   Error_Type = 16
	Error_ActualLRPCannotBeCrashed   Error_Type = 17
	Error_ActualLRPCannotBeFailed    Error_Type = 18
	Error_ActualLRPCannotBeRemoved   Error_Type = 19
	Error_ActualLRPCannotBeUnclaimed Error_Type = 21
	Error_RunningOnDifferentCell     Error_Type = 24
	Error_GUIDGeneration             Error_Type = 26
	Error_Deserialize                Error_Type = 27
	Error_Deadlock                   Error_Type = 28
	Error_Unrecoverable              Error_Type = 29
	Error_LockCollision              Error_Type = 30
	Error_Timeout                    Error_Type = 31
)

// Prevent copylock errors when using ProtoError directly
type Error struct {
	Type    Error_Type
	Message string
}

func (m *Error) GetType() Error_Type {
	if m != nil {
		return m.Type
	}
	return 0
}
func (m *Error) SetType(value Error_Type) {
	if m != nil {
		m.Type = value
	}
}
func (m *Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}
func (m *Error) SetMessage(value string) {
	if m != nil {
		m.Message = value
	}
}
func (x *Error) ToProto() *ProtoError {
	proto := &ProtoError{
		Type:    ProtoError_Type(x.Type),
		Message: x.Message,
	}
	return proto
}

func ErrorProtoMap(values []*Error) []*ProtoError {
	result := make([]*ProtoError, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}
