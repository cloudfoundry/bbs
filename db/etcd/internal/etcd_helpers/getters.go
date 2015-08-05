package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/storeadapter"

	. "github.com/onsi/gomega"
)

func (t *ETCDHelper) GetInstanceActualLRP(lrpKey *models.ActualLRPKey) (*models.ActualLRP, error) {
	resp, err := t.etcdClient.Get(etcd.ActualLRPSchemaPath(lrpKey.ProcessGuid, lrpKey.Index), false, false)
	if err == storeadapter.ErrorKeyNotFound {
		return &models.ActualLRP{}, models.ErrResourceNotFound
	}
	Expect(err).NotTo(HaveOccurred())

	var lrp models.ActualLRP
	err = models.FromJSON([]byte(resp.Node.Value), &lrp)
	Expect(err).NotTo(HaveOccurred())

	return &lrp, nil
}
