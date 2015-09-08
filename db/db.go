package db

//go:generate counterfeiter . DB
type DB interface {
	DomainDB
	LRPDB
	TaskDB
	EventDB
	EvacuationDB
	VersionDB
}
