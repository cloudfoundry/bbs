package db

//go:generate counterfeiter . DB

type DB interface {
	DomainDB
	EncryptionDB
	EvacuationDB
	LockDB
	LRPDB
	TaskDB
	VersionDB
}
