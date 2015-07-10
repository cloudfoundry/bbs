package bbs

import "github.com/gogo/protobuf/proto"

//go:generate protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. --gogofast_out=. error.proto
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
