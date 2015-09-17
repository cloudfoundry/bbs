package migrations

import (
	"errors"

	"github.com/cloudfoundry-incubator/bbs/db/deprecations"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
)

func init() {
	appendMigration(NewBase64ProtobufEncode())
}

type Base64ProtobufEncode struct {
	serializer format.Serializer
}

func NewBase64ProtobufEncode() Base64ProtobufEncode {
	return Base64ProtobufEncode{
		serializer: format.NewSerializer(nil),
	}
}

func (b Base64ProtobufEncode) Version() int64 {
	return 1441411196
}

func (b Base64ProtobufEncode) Up(logger lager.Logger, storeClient etcd.StoreClient) error {
	// Desired LRPs
	response, err := storeClient.Get(deprecations.DesiredLRPSchemaRoot, false, true)
	if err != nil {
		err = etcd.ErrorFromEtcdError(logger, err)

		// Continue if the root node does not exist
		if err != models.ErrResourceNotFound {
			return err
		}
	}

	if response != nil {
		desiredLRPRootNode := response.Node
		for _, node := range desiredLRPRootNode.Nodes {
			var desiredLRP models.DesiredLRP
			err := b.reWriteNode(logger, node, &desiredLRP, storeClient)
			if err != nil {
				return err
			}
		}
	}

	// Actual LRPs
	response, err = storeClient.Get(etcd.ActualLRPSchemaRoot, false, true)
	if err != nil {
		err = etcd.ErrorFromEtcdError(logger, err)

		// Continue if the root node does not exist
		if err != models.ErrResourceNotFound {
			return err
		}
	}

	if response != nil {
		actualLRPRootNode := response.Node
		for _, processNode := range actualLRPRootNode.Nodes {
			for _, groupNode := range processNode.Nodes {
				for _, actualLRPNode := range groupNode.Nodes {
					var actualLRP models.ActualLRP
					err := b.reWriteNode(logger, actualLRPNode, &actualLRP, storeClient)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Tasks
	response, err = storeClient.Get(etcd.TaskSchemaRoot, false, true)
	if err != nil {
		err = etcd.ErrorFromEtcdError(logger, err)

		// Continue if the root node does not exist
		if err != models.ErrResourceNotFound {
			return err
		}
	}

	if response != nil {
		taskRootNode := response.Node
		for _, node := range taskRootNode.Nodes {
			var task models.Task
			err := b.reWriteNode(logger, node, &task, storeClient)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b Base64ProtobufEncode) Down(logger lager.Logger, storeClient etcd.StoreClient) error {
	return errors.New("not implemented")
}

func (b Base64ProtobufEncode) reWriteNode(logger lager.Logger, node *goetcd.Node, model format.Versioner, storeClient etcd.StoreClient) error {
	err := b.serializer.Unmarshal(logger, []byte(node.Value), model)
	if err != nil {
		return err
	}

	value, err := b.serializer.Marshal(logger, format.ENCODED_PROTO, model)
	if err != nil {
		return err
	}

	_, err = storeClient.CompareAndSwap(node.Key, value, etcd.NO_TTL, node.ModifiedIndex)
	if err != nil {
		return err
	}

	return nil
}
