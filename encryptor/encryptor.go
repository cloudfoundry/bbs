package encryptor

import (
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db"
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

const (
	encryptionDuration = metric.Duration("EncryptionDuration")
)

type Encryptor struct {
	logger      lager.Logger
	db          db.DB
	keyManager  encryption.KeyManager
	cryptor     encryption.Cryptor
	storeClient etcddb.StoreClient
	clock       clock.Clock
}

func New(
	logger lager.Logger,
	db db.DB,
	keyManager encryption.KeyManager,
	cryptor encryption.Cryptor,
	storeClient etcddb.StoreClient,
	clock clock.Clock,
) Encryptor {
	return Encryptor{
		logger:      logger,
		db:          db,
		keyManager:  keyManager,
		cryptor:     cryptor,
		storeClient: storeClient,
		clock:       clock,
	}
}

func (m Encryptor) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := m.logger.Session("encryptor")

	currentEncryptionKey, err := m.db.EncryptionKeyLabel(logger)
	if err != nil {
		if models.ConvertError(err) != models.ErrResourceNotFound {
			return err
		}
	} else {
		if m.keyManager.DecryptionKey(currentEncryptionKey) == nil {
			return errors.New("Existing encryption key version (" + currentEncryptionKey + ") is not among the known keys")
		}
	}

	close(ready)

	if currentEncryptionKey != m.keyManager.EncryptionKey().Label() {
		encryptionStart := m.clock.Now()
		logger.Debug("encryption-started")
		m.performEncryption(logger)
		logger.Debug("encryption-finished")
		m.db.SetEncryptionKeyLabel(logger, m.keyManager.EncryptionKey().Label())
		encryptionDuration.Send(time.Since(encryptionStart))
	}

	select {
	case <-signals:
		return nil
	}
}

func (m Encryptor) performEncryption(logger lager.Logger) error {
	response, err := m.storeClient.Get(etcddb.V1SchemaRoot, false, true)
	if err != nil {
		err = etcddb.ErrorFromEtcdError(logger, err)

		// Continue if the root node does not exist
		if err != models.ErrResourceNotFound {
			return err
		}
	}

	if response != nil {
		rootNode := response.Node
		return m.rewriteNode(logger, rootNode)
	}

	return nil
}

func (m Encryptor) rewriteNode(logger lager.Logger, node *etcd.Node) error {
	if !node.Dir {
		encoder := format.NewEncoder(m.cryptor)
		payload, err := encoder.Decode([]byte(node.Value))
		if err != nil {
			logger.Error("failed-to-read-node", err, lager.Data{"etcd-key": node.Key})
			return nil
		}
		encryptedPayload, err := encoder.Encode(format.BASE64_ENCRYPTED, payload)
		if err != nil {
			return err
		}
		_, err = m.storeClient.CompareAndSwap(node.Key, encryptedPayload, etcddb.NO_TTL, node.ModifiedIndex)
		if err != nil {
			logger.Info("failed-to-compare-and-swap", lager.Data{"err": err, "etcd-key": node.Key})
			return nil
		}
	} else {
		for _, child := range node.Nodes {
			err := m.rewriteNode(logger, child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
