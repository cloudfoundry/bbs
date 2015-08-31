package models

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
)

type SerializationFormat byte

const JSON_NO_ENVELOPE SerializationFormat = 0
const JSON SerializationFormat = 1
const PROTO SerializationFormat = 2

type Version byte

const V0 Version = 0

type Envelope struct {
	SerializationFormat SerializationFormat
	Version             Version
	Payload             []byte
}

func OpenEnvelope(data []byte) *Envelope {
	e := &Envelope{}

	if !isEnveloped(data) {
		e.SerializationFormat = JSON
		e.Version = V0
		e.Payload = data
		return e
	}

	e.SerializationFormat = SerializationFormat(data[0])
	e.Version = Version(data[1])
	e.Payload = data[2:]
	return e
}

func MarshalEnvelope(format SerializationFormat, model Versioner) ([]byte, *Error) {
	var payload []byte
	var err *Error

	switch format {
	case PROTO:
		payload, err = toProto(model)
	case JSON:
		payload, err = ToJSON(model)
	case JSON_NO_ENVELOPE:
		return ToJSON(model)
	default:
		err = NewError(Error_InvalidRecord, "unknown format")
	}

	if err != nil {
		return nil, err
	}

	// to avoid the following copy, change toProto to write the payload
	// into a buffer pre-filled with format and version.
	data := make([]byte, len(payload)+2)
	data[0] = byte(format)
	data[1] = byte(model.Version())
	for i := range payload {
		data[i+2] = payload[i]
	}

	return data, nil
}

func (e *Envelope) Unmarshal(logger lager.Logger, model Versioner) *Error {
	switch e.SerializationFormat {
	case JSON:
		err := json.Unmarshal(e.Payload, model)
		if err != nil {
			logger.Error("failed-to-json-unmarshal-payload", err)
			return NewError(Error_InvalidRecord, err.Error())
		}
	case PROTO:
		err := proto.Unmarshal(e.Payload, model)
		if err != nil {
			logger.Error("failed-to-proto-unmarshal-payload", err)
			return NewError(Error_InvalidRecord, err.Error())
		}
	default:
		logger.Error("cannot-unmarshal-unknown-serialization-format", nil)
		return NewError(Error_FailedToOpenEnvelope, "unknown serialization format")
	}

	model.MigrateFromVersion(e.Version)

	err := model.Validate()
	if err != nil {
		logger.Error("invalid-record", err)
		return NewError(Error_InvalidRecord, err.Error())
	}
	return nil
}

func isEnveloped(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	switch SerializationFormat(data[0]) {
	case JSON, PROTO:
	default:
		return false
	}

	switch Version(data[1]) {
	case V0:
	default:
		return false
	}

	return true
}
