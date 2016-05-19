package format

import "github.com/gogo/protobuf/proto"

type Version byte

const (
	V0 Version = 0
	V1         = 1
)

var ValidVersions = []Version{V0, V1}

//go:generate counterfeiter . Versioner
type Versioner interface {
	MigrateFromVersion(v Version) error
	Validate() error
	Version() Version
}

//go:generate counterfeiter . ProtoVersioner
type ProtoVersioner interface {
	proto.Message
	Versioner
}
