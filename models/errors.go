package models

import "errors"

type ErrInvalidParameter struct {
	Parameter string
}

func (err ErrInvalidParameter) Error() string {
	return "Invalid parameter: " + err.Parameter
}

type ErrInvalidField struct {
	Field string
}

func (err ErrInvalidField) Error() string {
	return "Invalid field: " + err.Field
}

type ErrInvalidModification struct {
	InvalidField string
}

func (err ErrInvalidModification) Error() string {
	return "attempt to make invalid change to field: " + err.InvalidField
}

var ErrActualLRPGroupInvalid = errors.New("ActualLRPGroup invalid")
