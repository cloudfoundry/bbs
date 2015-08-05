package db

type DB interface {
	DomainDB
	ActualLRPDB
	DesiredLRPDB
	TaskDB
	EventDB
	EvacuationDB
}
