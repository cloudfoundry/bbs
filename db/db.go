package db

//go:generate counterfeiter . DB

type DB interface {
	DomainDB
	EncryptionDB
	EvacuationDB
	EventDB
	LRPDB
	TaskDB
	VersionDB
}
