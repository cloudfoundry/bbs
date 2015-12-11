package etcd_test

import (
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StoreClient", func() {
	Describe("Create", func() {
		It("allows duplicate keys if contents are the same", func() {
			_, err := storeClient.Create("a", []byte("thing"), 0)
			Expect(err).NotTo(HaveOccurred())

			_, err = storeClient.Create("a", []byte("thing"), 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not allows duplicate keys if contents are different", func() {
			_, err := storeClient.Create("a", []byte("thing"), 0)
			Expect(err).NotTo(HaveOccurred())

			_, err = storeClient.Create("a", []byte("thing-2"), 0)
			Expect(err).To(HaveOccurred())
			Expect(err.(*etcd.EtcdError).ErrorCode).To(Equal(etcddb.ETCDErrKeyExists))
		})
	})
})
