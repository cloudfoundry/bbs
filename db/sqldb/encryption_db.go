package sqldb

import "github.com/pivotal-golang/lager"

const EncryptionKeyID = "encryption_key_label"

func (db *SQLDB) SetEncryptionKeyLabel(logger lager.Logger, label string) error {
	return db.setConfigurationValue(logger, EncryptionKeyID, label)
}

func (db *SQLDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
	return db.getConfigurationValue(logger, EncryptionKeyID)
}
