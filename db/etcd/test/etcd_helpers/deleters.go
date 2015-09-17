package etcd_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	. "github.com/onsi/gomega"
)

func (t *ETCDHelper) DeleteDesiredLRP(guid string) {
	_, err := t.client.Delete(etcd.DesiredLRPSchedulingInfoSchemaPath(guid), false)
	Expect(err).NotTo(HaveOccurred())
	_, err = t.client.Delete(etcd.DesiredLRPRunInfoSchemaPath(guid), false)
	Expect(err).NotTo(HaveOccurred())
}
