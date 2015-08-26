package models

import (
	"encoding/json"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

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

func toProto(v ProtoValidator) ([]byte, *Error) {
	if !isNil(v) {
		if err := v.Validate(); err != nil {
			return nil, NewError(InvalidRecord, err.Error())
		}
	}

	bytes, err := proto.Marshal(v)
	if err != nil {
		return nil, NewError(InvalidProtobufMessage, err.Error())
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
