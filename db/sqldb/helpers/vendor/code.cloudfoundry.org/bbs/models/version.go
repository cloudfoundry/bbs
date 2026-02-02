package models

// SchemaVersion holds the current database migration version (used by DB/migrations).
type SchemaVersion struct {
	CurrentVersion int64
}
