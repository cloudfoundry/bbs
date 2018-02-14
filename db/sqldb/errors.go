package sqldb

import (
	"bytes"
	"fmt"

	"code.cloudfoundry.org/bbs/models"
)

type dbErr struct {
	ops   []string
	fatal bool
	err   error
}

func (e dbErr) Error() string {
	var buf bytes.Buffer
	for idx, op := range e.ops {
		if idx > 0 {
			fmt.Fprint(&buf, " / ")
		}
		fmt.Fprint(&buf, op)
	}
	fmt.Fprintf(&buf, ": %s", e.err)
	return buf.String()
}

// Fatal or unexpected errror
func F(op string, err error) error {
	e := E(op, err)
	newE := e.(dbErr)
	newE.fatal = true
	return newE
}

// Non-fatal error
func E(op string, err error) error {
	e, ok := err.(dbErr)
	if !ok {
		return dbErr{ops: []string{op}, err: err}
	}
	e.ops = append([]string{op}, e.ops...)
	return e
}

// true if this is a fatal or unexpected error. Otherwise, true if this is an
// avoidable error, e.g. row not found or duplicate row was inserted
// (ErrResourceNotFound and ErrResourceExists, resp.).
func IsAvoidableE(err error) bool {
	e, ok := err.(dbErr)
	if !ok {
		return false
	}
	return !e.fatal && (e.err == models.ErrResourceExists || e.err == models.ErrResourceNotFound)
}
