package etcd_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/codec"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	etcdclient "github.com/coreos/go-etcd/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StoreClient", func() {
	var (
		encoding    codec.Kind
		storeClient etcd.StoreClient
	)

	const (
		jsonMap  = `{"key":"value"}`
		jsonList = `["key1", "key2"]`
	)

	basicSetup := func() {
		etcdClient.Set("/valid//none/json-map", jsonMap, 0)
		etcdClient.Set("/valid/none/json-list", jsonList, 0)
		etcdClient.Set("/valid/unencoded/json-map", "00"+jsonMap, 0)
		etcdClient.Set("/valid/unencoded/json-list", "00"+jsonList, 0)
		etcdClient.Set("/valid/base64/json-map", `01eyJrZXkiOiJ2YWx1ZSJ9`, 0)
		etcdClient.Set("/valid/base64/json-list", `01WyJrZXkxIiwgImtleTIiXQ==`, 0)
		etcdClient.Set("/invalid/one/header", `99Garbage`, 0)
		etcdClient.Set("/invalid/two/encoding", `01$$$$$$$$`, 0)
	}

	BeforeEach(func() {
		encoding = codec.NONE
	})

	JustBeforeEach(func() {
		storeClient = etcd.NewStoreClient(etcdClient, encoding)
	})

	AfterEach(func() {
		storeClient = nil
	})

	getOne := func(key string, value string) {
		response, err := storeClient.Get(key, false, false)
		if value != "" {
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Node.Value).To(Equal(value))
		} else {
			Expect(err).To(HaveOccurred())
		}
	}

	Describe("Get", func() {
		BeforeEach(func() {
			basicSetup()
		})

		Context("when the data in the store is valid", func() {
			It("should be able to decode valid data", func() {
				getOne("/valid/none/json-map", jsonMap)
				getOne("/valid/none/json-list", jsonList)
				getOne("/valid/unencoded/json-map", jsonMap)
				getOne("/valid/unencoded/json-list", jsonList)
				getOne("/valid/base64/json-map", jsonMap)
				getOne("/valid/base64/json-list", jsonList)
			})
		})

		Context("when the data in the store is not valid", func() {
			It("should return an error", func() {
				getOne("/invalid/one/header", "")
				getOne("/invalid/two/encoding", "")
			})
		})

		Context("when retrieving recursively", func() {
			Context("and the data in the store is valid", func() {
				It("should be able to decode the data", func() {
					response, err := storeClient.Get("/valid", true, true)
					Expect(err).NotTo(HaveOccurred())

					Expect(collect(response.Node, nil)).To(Equal(map[string]string{
						"/valid":                     "",
						"/valid/none":                "",
						"/valid/none/json-map":       jsonMap,
						"/valid/none/json-list":      jsonList,
						"/valid/unencoded":           "",
						"/valid/unencoded/json-map":  jsonMap,
						"/valid/unencoded/json-list": jsonList,
						"/valid/base64":              "",
						"/valid/base64/json-map":     jsonMap,
						"/valid/base64/json-list":    jsonList,
					}))
				})
			})

			Context("and the data in the store is not valid", func() {
				It("should raise an error", func() {
					_, err := storeClient.Get("/invalid", true, true)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func collect(node *etcdclient.Node, tree map[string]string) map[string]string {
	if tree == nil {
		tree = map[string]string{}
	}

	tree[node.Key] = node.Value
	for _, n := range node.Nodes {
		collect(n, tree)
	}

	return tree
}
