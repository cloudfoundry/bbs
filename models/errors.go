package models

import "errors"

func NewError(errType string, msg string) *Error {
	return &Error{
		Type:    errType,
		Message: msg,
	}
}

func (err Error) Error() string {
	return err.GetMessage()
}

const (
	InvalidDomain = "InvalidDomain"

	InvalidRecord          = "InvalidRecord"
	InvalidRequest         = "InvalidRequest"
	InvalidResponse        = "InvalidResponse"
	InvalidProtobufMessage = "InvalidProtobufMessage"
	InvalidJSON            = "InvalidJSON"
	InvalidStateTransition = "InvalidStateTransition"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ResourceConflict = "ResourceConflict"
	ResourceExists   = "ResourceExists"
	ResourceNotFound = "ResourceNotFound"
	RouterError      = "RouterError"

	ActualLRPCannotBeClaimed   = "ActualLRPCannotBeClaimed"
	ActualLRPCannotBeStarted   = "ActualLRPCannotBeStarted"
	ActualLRPCannotBeCrashed   = "ActualLRPCannotBeCrashed"
	ActualLRPCannotBeFailed    = "ActualLRPCannotBeFailed"
	ActualLRPCannotBeRemoved   = "ActualLRPCannotBeRemoved"
	ActualLRPCannotBeStopped   = "ActualLRPCannotBeStopped"
	ActualLRPCannotBeUnclaimed = "ActualLRPCannotBeUnclaimed"
	ActualLRPCannotBeEvacuated = "ActualLRPCannotBeEvacuated"
)

var (
	ErrResourceNotFound = &Error{
		Type:    ResourceNotFound,
		Message: "the requested resource could not be found",
	}

	ErrResourceExists = &Error{
		Type:    ResourceExists,
		Message: "the requested resource already exists",
	}

	ErrResourceConflict = &Error{
		Type:    ResourceConflict,
		Message: "the requested resource is in a conflicting state",
	}

	ErrBadRequest = &Error{
		Type:    InvalidRequest,
		Message: "the request received is invalid",
	}

	ErrUnknownError = &Error{
		Type:    UnknownError,
		Message: "the request failed for an unknown reason",
	}

	ErrSerializeJSON = &Error{
		Type:    InvalidJSON,
		Message: "could not serialize JSON",
	}

	ErrDeserializeJSON = &Error{
		Type:    InvalidJSON,
		Message: "could not deserialize JSON",
	}

	ErrActualLRPCannotBeClaimed = &Error{
		Type:    ActualLRPCannotBeClaimed,
		Message: "cannot claim actual LRP",
	}

	ErrActualLRPCannotBeStarted = &Error{
		Type:    ActualLRPCannotBeStarted,
		Message: "cannot start actual LRP",
	}

	ErrActualLRPCannotBeCrashed = &Error{
		Type:    ActualLRPCannotBeCrashed,
		Message: "cannot crash actual LRP",
	}

	ErrActualLRPCannotBeFailed = &Error{
		Type:    ActualLRPCannotBeFailed,
		Message: "cannot fail actual LRP",
	}

	ErrActualLRPCannotBeRemoved = &Error{
		Type:    ActualLRPCannotBeRemoved,
		Message: "cannot remove actual LRP",
	}

	ErrActualLRPCannotBeStopped = &Error{
		Type:    ActualLRPCannotBeStopped,
		Message: "cannot stop actual LRP",
	}

	ErrActualLRPCannotBeUnclaimed = &Error{
		Type:    ActualLRPCannotBeUnclaimed,
		Message: "cannot unclaim actual LRP",
	}

	ErrActualLRPCannotBeEvacuated = &Error{
		Type:    ActualLRPCannotBeEvacuated,
		Message: "cannot evacuate actual LRP",
	}
)

func (err *Error) Equal(other error) bool {
	if e, ok := other.(*Error); ok {
		return e.GetType() == err.GetType()
	}
	return false
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
