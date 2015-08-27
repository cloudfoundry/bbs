package db

type DB interface {
	DomainDB
	LRPDB
	TaskDB
	EventDB
	EvacuationDB
}
