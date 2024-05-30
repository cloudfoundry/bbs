// Code generated by protoc-gen-go-bbs. DO NOT EDIT.
// versions:
// - protoc-gen-go-bbs v0.0.1
// - protoc            v5.27.0
// source: error.proto

package models

import (
	strconv "strconv"
)

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

// Enum value maps for Error_Type
var (
	Error_Type_name = map[int32]string{
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
	Error_Type_value = map[string]int32{
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

func (m Error_Type) String() string {
	s, ok := Error_Type_name[int32(m)]
	if ok {
		return s
	}
	return strconv.Itoa(int(m))
}

// Prevent copylock errors when using ProtoError directly
type Error struct {
	Type    Error_Type `json:"type"`
	Message string     `json:"message"`
}

func (this *Error) Equal(that interface{}) bool {

	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Error)
	if !ok {
		that2, ok := that.(Error)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}

	if this.Type != that1.Type {
		return false
	}
	return true
}
func (m *Error) GetType() Error_Type {
	if m != nil {
		return m.Type
	}
	var defaultValue Error_Type
	defaultValue = 0
	return defaultValue
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
	var defaultValue string
	defaultValue = ""
	return defaultValue
}
func (m *Error) SetMessage(value string) {
	if m != nil {
		m.Message = value
	}
}
func (x *Error) ToProto() *ProtoError {
	if x == nil {
		return nil
	}

	proto := &ProtoError{
		Type:    ProtoError_Type(x.Type),
		Message: x.Message,
	}
	return proto
}

func (x *ProtoError) FromProto() *Error {
	if x == nil {
		return nil
	}

	copysafe := &Error{
		Type:    Error_Type(x.Type),
		Message: x.Message,
	}
	return copysafe
}

func ErrorToProtoSlice(values []*Error) []*ProtoError {
	if values == nil {
		return nil
	}
	result := make([]*ProtoError, len(values))
	for i, val := range values {
		result[i] = val.ToProto()
	}
	return result
}

func ErrorFromProtoSlice(values []*ProtoError) []*Error {
	if values == nil {
		return nil
	}
	result := make([]*Error, len(values))
	for i, val := range values {
		result[i] = val.FromProto()
	}
	return result
}
