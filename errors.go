package bbs

type Error struct {
	Type    string `json:"name"`
	Message string `json:"message"`
}

func (err Error) Error() string {
	return err.Message
}

const (
	InvalidDomain = "InvalidDomain"

	InvalidJSON     = "InvalidJSON"
	InvalidRequest  = "InvalidRequest"
	InvalidResponse = "InvalidResponse"

	UnknownError = "UnknownError"
	Unauthorized = "Unauthorized"

	ActualLRPIndexNotFound = "ActualLRPIndexNotFound"

	ResourceConflict = "ResourceConflict"
	RouterError      = "RouterError"
)
