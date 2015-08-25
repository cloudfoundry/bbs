package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	etcdclient "github.com/coreos/go-etcd/etcd"

	. "github.com/onsi/gomega"
)

func (t *ETCDHelper) GetInstanceActualLRP(lrpKey *models.ActualLRPKey) (*models.ActualLRP, error) {
	resp, err := t.etcdClient.Get(etcd.ActualLRPSchemaPath(lrpKey.ProcessGuid, lrpKey.Index), false, false)
	if etcdErr, ok := err.(*etcdclient.EtcdError); ok && etcdErr.ErrorCode == etcd.ETCDErrKeyNotFound {
		return &models.ActualLRP{}, models.ErrResourceNotFound
	}

	Expect(err).NotTo(HaveOccurred())

	var lrp models.ActualLRP
	err = models.FromJSON([]byte(resp.Node.Value), &lrp)
	Expect(err).NotTo(HaveOccurred())

	return &lrp, nil
}
