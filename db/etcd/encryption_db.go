package etcd

import "github.com/pivotal-golang/lager"

func (db *ETCDDB) SetEncryptionKeyLabel(logger lager.Logger, keyLabel string) error {
	logger.Debug("set-encryption-key-label", lager.Data{"encryption-key-label": keyLabel})
	defer logger.Debug("set-encryption-key-label-finished")

	_, err := db.client.Set(EncryptionKeyLabelKey, []byte(keyLabel), NO_TTL)
	return err
}

func (db *ETCDDB) EncryptionKeyLabel(logger lager.Logger) (string, error) {
	logger.Debug("get-encryption-key-label")
	defer logger.Debug("get-encryption-key-label-finished")

	node, err := db.fetchRaw(logger, EncryptionKeyLabelKey)
	if err != nil {
		return "", err
	}

	return node.Value, nil
}
