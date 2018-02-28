package format

import "github.com/gogo/protobuf/proto"

type Version byte

const (
	V0 Version = 0
	V1         = 1
	V2         = 2
)

var ValidVersions = []Version{V0, V1, V2}

type Model interface {
	proto.Message
}
