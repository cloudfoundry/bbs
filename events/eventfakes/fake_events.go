package eventfakes

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type FakeEvent struct{ Token string }

func (FakeEvent) EventType() string           { return "fake" }
func (FakeEvent) Key() string                 { return "fake" }
func (FakeEvent) ProtoMessage()               {}
func (FakeEvent) ToEventProto() proto.Message { return nil }
func (FakeEvent) Reset()                      {}
func (FakeEvent) String() string              { return "fake" }
func (e FakeEvent) Marshal() ([]byte, error)  { return []byte(e.Token), nil }

type UnmarshalableEvent struct{ Fn func() }

func (UnmarshalableEvent) EventType() string           { return "unmarshalable" }
func (UnmarshalableEvent) Key() string                 { return "unmarshalable" }
func (UnmarshalableEvent) ProtoMessage()               {}
func (UnmarshalableEvent) ToEventProto() proto.Message { return nil }
func (UnmarshalableEvent) Reset()                      {}
func (UnmarshalableEvent) String() string              { return "unmarshalable" }
func (UnmarshalableEvent) Marshal() ([]byte, error)    { return nil, errors.New("no workie") }
