package fakesqldriver

import "database/sql/driver"

//go:generate counterfeiter -generate

//counterfeiter:generate . Driver
type Driver interface {
	Open(name string) (driver.Conn, error)
}

//counterfeiter:generate . Conn
type Conn interface {
	Prepare(query string) (driver.Stmt, error)
	Close() error
	Begin() (driver.Tx, error)
}

//counterfeiter:generate . Tx
type Tx interface {
	Commit() error
	Rollback() error
}

//counterfeiter:generate . Stmt
type Stmt interface {
	Close() error
	NumInput() int
	Exec(args []driver.Value) (driver.Result, error)
	Query(args []driver.Value) (driver.Rows, error)
}
