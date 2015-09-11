package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/gomega"
)

func (t *ETCDHelper) GetInstanceActualLRP(lrpKey *models.ActualLRPKey) (*models.ActualLRP, error) {
	resp, err := t.client.Get(etcd.ActualLRPSchemaPath(lrpKey.ProcessGuid, lrpKey.Index), false, false)
	if etcdErr, ok := err.(*etcdclient.EtcdError); ok && etcdErr.ErrorCode == etcd.ETCDErrKeyNotFound {
		return &models.ActualLRP{}, models.ErrResourceNotFound
	}

	Expect(err).NotTo(HaveOccurred())

	var lrp models.ActualLRP
	logger := lager.NewLogger("etcd-helper")
	err = format.Unmarshal(logger, []byte(resp.Node.Value), &lrp)
	Expect(err).NotTo(HaveOccurred())

	return &lrp, nil
}
