package bbs

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
