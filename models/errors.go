package models

import (
	"errors"

	"github.com/gogo/protobuf/proto"
)

func (err Error) Error() string {
	return err.GetMessage()
}

const (
	InvalidDomain = "InvalidDomain"

	InvalidRequest         = "InvalidRequest"
	InvalidResponse        = "InvalidResponse"
	InvalidProtobufMessage = "InvalidProtobufMessage"
	InvalidJSON            = "InvalidJSON"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ResourceConflict = "ResourceConflict"
	ResourceNotFound = "ResourceNotFound"
	RouterError      = "RouterError"
)

var (
	ErrResourceNotFound = &Error{
		Type:    proto.String(ResourceNotFound),
		Message: proto.String("the requested resource could not be found"),
	}

	ErrUnknownError = &Error{
		Type:    proto.String(UnknownError),
		Message: proto.String("the request failed for an unknown reason"),
	}

	ErrDeserializeJSON = &Error{
		Type:    proto.String(InvalidJSON),
		Message: proto.String("could not deserialize JSON"),
	}
)

func (err *Error) Equal(other *Error) bool {
	return err.GetType() == other.GetType()
}

type ErrInvalidField struct {
	Field string
}

func (err ErrInvalidField) Error() string {
	return "Invalid field: " + err.Field
}

type ErrInvalidModification struct {
	InvalidField string
}

func (err ErrInvalidModification) Error() string {
	return "attempt to make invalid change to field: " + err.InvalidField
}

var ErrActualLRPGroupInvalid = errors.New("ActualLRPGroup invalid")
