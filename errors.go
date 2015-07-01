package bbs

//go:generate protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. --gogofast_out=. error.proto
func (err Error) Error() string {
	return err.GetMessage()
}

const (
	InvalidDomain = "InvalidDomain"

	InvalidRequest         = "InvalidRequest"
	InvalidResponse        = "InvalidResponse"
	InvalidProtobufMessage = "InvalidProtobufMessage"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ResourceConflict = "ResourceConflict"
	RouterError      = "RouterError"
)
