package models

import (
	"encoding/json"
	"reflect"
)

type Payload struct {
	Version Version
	Payload []byte
}

type Version [4]byte

var (
	V0 Version = [4]byte{'0', '0', '0', '0'}
	V1 Version = [4]byte{'0', '0', '0', '1'}
)

func NewPayload(payload []byte) (*Payload, error) {
	version := V0
	if len(payload) >= len(V0) {
		version = Version([4]byte{payload[0], payload[1], payload[2], payload[3]})
	}

	switch version {
	case V1:
		return &Payload{
			Version: version,
			Payload: payload[4:],
		}, nil
	default:
		return &Payload{
			Version: V0,
			Payload: payload,
		}, nil
	}
}

func FromJSON(payload []byte, v Validator) error {
	err := json.Unmarshal(payload, v)
	if err != nil {
		return err
	}
	return v.Validate()
}

func ToJSON(v Validator) ([]byte, *Error) {
	if !isNil(v) {
		if err := v.Validate(); err != nil {
			return nil, NewError(InvalidRecord, err.Error())
		}
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, NewError(InvalidJSON, err.Error())
	}

	return bytes, nil
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}

	switch reflect.TypeOf(a).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(a).IsNil()
	}

	return false
}
