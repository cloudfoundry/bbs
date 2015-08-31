package models

import (
	"bytes"

	"github.com/gogo/protobuf/proto"
)

//go:generate counterfeiter . Validator
type Validator interface {
	Validate() error
}

//go:generate counterfeiter . ProtoValidator
type ProtoValidator interface {
	Validator
	proto.Message
}

type ValidationError []error

func (ve ValidationError) Append(err error) ValidationError {
	switch err := err.(type) {
	case ValidationError:
		return append(ve, err...)
	default:
		return append(ve, err)
	}
}

func (ve ValidationError) Error() string {
	var buffer bytes.Buffer

	for i, err := range ve {
		if err == nil {
			continue
		}
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(err.Error())
	}

	return buffer.String()
}

func (ve ValidationError) Empty() bool {
	return len(ve) == 0
}
