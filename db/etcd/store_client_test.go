package etcd_test

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/db/codec"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	etcderror "github.com/coreos/etcd/error"
	etcdclient "github.com/coreos/go-etcd/etcd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	jsonMap  = `{"key":"value"}`
	jsonList = `["key1", "key2"]`
)

var (
	validEncodedData = map[string]string{
		"/valid/none/json-map":       jsonMap,
		"/valid/none/json-list":      jsonList,
		"/valid/unencoded/json-map":  "00" + jsonMap,
		"/valid/unencoded/json-list": "00" + jsonList,
		"/valid/base64/json-map":     `01eyJrZXkiOiJ2YWx1ZSJ9`,
		"/valid/base64/json-list":    `01WyJrZXkxIiwgImtleTIiXQ==`,
	}

	validDecodedData = map[string]string{
		"/valid/none/json-map":       jsonMap,
		"/valid/none/json-list":      jsonList,
		"/valid/unencoded/json-map":  jsonMap,
		"/valid/unencoded/json-list": jsonList,
		"/valid/base64/json-map":     jsonMap,
		"/valid/base64/json-list":    jsonList,
	}

	invalidEncodedData = map[string]string{
		"/invalid/one/header":   `99Garbage`,
		"/invalid/two/encoding": `01$$$$$$$$`,
	}
)

var _ = Describe("StoreClient", func() {
	var encoding codec.Kind
	var storeClient etcd.StoreClient
	var etcdClient *etcdclient.Client

	BeforeEach(func() {
		encoding = codec.NONE
		etcdClient = etcdRunner.Client()
		etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	})

	JustBeforeEach(func() {
		storeClient = etcd.NewStoreClient(etcdClient, encoding)
	})

	AfterEach(func() {
		storeClient = nil
	})

	Describe("Get", func() {
		BeforeEach(func() {
			basicSetup(etcdClient)
		})

		Context("when the data in the store is valid", func() {
			It("should be able to decode valid data", func() {
				for key, value := range validDecodedData {
					response, err := storeClient.Get(key, false, false)
					Expect(err).NotTo(HaveOccurred())
					Expect(response.Node.Value).To(Equal(value))
				}
			})
		})

		Context("when the data in the store is not valid", func() {
			It("should return an error", func() {
				for key, _ := range invalidEncodedData {
					_, err := storeClient.Get(key, false, false)
					Expect(err).To(HaveOccurred())
				}
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

	Describe("Set", func() {
		Context("when using codec.NONE", func() {
			BeforeEach(func() {
				encoding = codec.NONE
			})

			It("stores the payload exactly as presented", func() {
				_, err := storeClient.Set("/valid/none/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/none/json-map", false, false)
				Expect(response.Node.Value).To(Equal(jsonMap))
			})
		})

		Context("when using codec.UNENCODED", func() {
			BeforeEach(func() {
				encoding = codec.UNENCODED
			})

			It("stores the payload with the correct Kind prefix", func() {
				_, err := storeClient.Set("/valid/unencoded/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/unencoded/json-map", false, false)
				Expect(response.Node.Value).To(Equal(validEncodedData[response.Node.Key]))
			})
		})

		Context("when using codec.BASE64", func() {
			BeforeEach(func() {
				encoding = codec.BASE64
			})

			It("stores the payload with the correct encoding and Kind prefix", func() {
				_, err := storeClient.Set("/valid/base64/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/base64/json-map", false, false)
				Expect(response.Node.Value).To(Equal(validEncodedData[response.Node.Key]))
			})
		})
	})

	Describe("Create", func() {
		Context("when using codec.NONE", func() {
			BeforeEach(func() {
				encoding = codec.NONE
			})

			It("stores the payload exactly as presented", func() {
				_, err := storeClient.Create("/valid/none/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/none/json-map", false, false)
				Expect(response.Node.Value).To(Equal(jsonMap))
			})
		})

		Context("when using codec.UNENCODED", func() {
			BeforeEach(func() {
				encoding = codec.UNENCODED
			})

			It("stores the payload with the correct Kind prefix", func() {
				_, err := storeClient.Create("/valid/unencoded/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/unencoded/json-map", false, false)
				Expect(response.Node.Value).To(Equal(validEncodedData[response.Node.Key]))
			})
		})

		Context("when using codec.BASE64", func() {
			BeforeEach(func() {
				encoding = codec.BASE64
			})

			It("stores the payload with the correct encoding and Kind prefix", func() {
				_, err := storeClient.Create("/valid/base64/json-map", []byte(jsonMap), 0)
				Expect(err).NotTo(HaveOccurred())

				response, err := etcdClient.Get("/valid/base64/json-map", false, false)
				Expect(response.Node.Value).To(Equal(validEncodedData[response.Node.Key]))
			})
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			basicSetup(etcdClient)
		})

		Context("when recursive is false", func() {
			It("deletes the the specified key", func() {
				_, err := storeClient.Delete("/valid/none/json-map", false)
				Expect(err).NotTo(HaveOccurred())

				result, err := storeClient.Get("/valid/none/json-list", false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Node.Value).To(Equal(jsonList))

				_, err = storeClient.Get("/valid/none/json-map", false, false)
				Expect(err).To(HaveOccurred())
				etcdErr := err.(*etcdclient.EtcdError)
				Expect(etcdErr.ErrorCode).To(Equal(etcderror.EcodeKeyNotFound))
			})
		})

		Context("when recursive is true", func() {
			It("deletes the specified node recursively", func() {
				_, err := storeClient.Delete("/valid/none", true)
				Expect(err).NotTo(HaveOccurred())

				_, err = storeClient.Get("/valid/none/json-map", false, false)
				Expect(err).To(HaveOccurred())
				etcdErr := err.(*etcdclient.EtcdError)
				Expect(etcdErr.ErrorCode).To(Equal(etcderror.EcodeKeyNotFound))

				_, err = storeClient.Get("/valid/none/json-list", false, false)
				Expect(err).To(HaveOccurred())
				etcdErr = err.(*etcdclient.EtcdError)
				Expect(etcdErr.ErrorCode).To(Equal(etcderror.EcodeKeyNotFound))
			})
		})
	})

	Describe("CompareAndSwap", func() {
		var oldIndex uint64

		BeforeEach(func() {
			basicSetup(etcdClient)
			encoding = codec.UNENCODED

			response, err := etcdClient.Get("/valid/unencoded/json-map", false, false)
			Expect(err).NotTo(HaveOccurred())
			oldIndex = response.Node.ModifiedIndex
		})

		Context("when compare and swap is successful", func() {
			It("returns unencoded values in the response", func() {
				response, err := storeClient.CompareAndSwap("/valid/unencoded/json-map", []byte(`{"new":"value"}`), 0, oldIndex)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Node.ModifiedIndex).To(BeNumerically(">", oldIndex))
				Expect(response.Node.Value).To(BeEmpty())
				Expect(response.PrevNode.ModifiedIndex).To(Equal(oldIndex))
				Expect(response.PrevNode.Value).To(BeEmpty())
			})
		})

		Context("when compare and swap fails", func() {
			BeforeEach(func() {
				_, err := etcdClient.Set("/valid/unencoded/json-map", jsonMap, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an etcd comparison failed ('test') error", func() {
				_, err := storeClient.CompareAndSwap("/valid/unencoded/json-map", []byte(`{"new":"value"}`), 0, oldIndex)
				Expect(err).To(HaveOccurred())
				etcdErr := err.(*etcdclient.EtcdError)
				Expect(etcdErr.ErrorCode).To(Equal(etcderror.EcodeTestFailed))
			})
		})
	})

	Describe("CompareAndDelete", func() {
		var oldIndex uint64

		BeforeEach(func() {
			basicSetup(etcdClient)
			encoding = codec.UNENCODED

			response, err := etcdClient.Get("/valid/unencoded/json-map", false, false)
			Expect(err).NotTo(HaveOccurred())
			oldIndex = response.Node.ModifiedIndex
		})

		Context("when compare and delete is successful", func() {
			It("returns unencoded values in the response", func() {
				response, err := storeClient.CompareAndDelete("/valid/unencoded/json-map", oldIndex)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Node.ModifiedIndex).To(BeNumerically(">", oldIndex))
				Expect(response.Node.Value).To(BeEmpty())
				Expect(response.PrevNode.ModifiedIndex).To(Equal(oldIndex))
				Expect(response.PrevNode.Value).To(BeEmpty())
			})
		})

		Context("when compare and delete fails", func() {
			BeforeEach(func() {
				_, err := etcdClient.Set("/valid/unencoded/json-map", jsonMap, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an etcd comparison failed ('test') error", func() {
				_, err := storeClient.CompareAndDelete("/valid/unencoded/json-map", oldIndex)
				Expect(err).To(HaveOccurred())
				etcdErr := err.(*etcdclient.EtcdError)
				Expect(etcdErr.ErrorCode).To(Equal(etcderror.EcodeTestFailed))
			})
		})
	})

	Describe("Watch", func() {
		var (
			updates  chan *etcdclient.Response
			stop     chan bool
			complete chan error
		)

		BeforeEach(func() {
			basicSetup(etcdClient)
		})

		Context("when the watch is a one-shot", func() {
			It("returns the first change", func() {
				responseChan := make(chan *etcdclient.Response)

				go func() {
					defer GinkgoRecover()
					response, watchErr := storeClient.Watch("/valid", 0, true, nil, nil)
					Expect(watchErr).NotTo(HaveOccurred())
					responseChan <- response
				}()

				// laaaaaame :(
				time.Sleep(time.Millisecond)

				key := "/valid/base64/json-map"
				_, err := etcdClient.Set(key, validEncodedData[key], 0)
				Expect(err).NotTo(HaveOccurred())

				response := <-responseChan
				Expect(response.PrevNode.Key).To(Equal(key))
				Expect(response.PrevNode.Value).To(Equal(validDecodedData[key]))
				Expect(response.Node.Key).To(Equal(key))
				Expect(response.Node.Value).To(Equal(validDecodedData[key]))
			})
		})

		Context("when a receiver is provided", func() {
			JustBeforeEach(func() {
				updates = make(chan *etcdclient.Response, 1)
				stop = make(chan bool)
				complete = make(chan error)

				go func() {
					_, err := storeClient.Watch("/valid", 0, true, updates, stop)
					complete <- err
				}()

				Eventually(func() bool {
					storeClient.Set("/valid/setup", []byte("value"), 0)

					select {
					case <-updates:
						return true
					default:
						return false
					}
				}).Should(BeTrue())
			})

			AfterEach(func() {
				close(stop)
				Eventually(complete).Should(Receive())
				Eventually(updates).Should(BeClosed())
			})

			It("streams the unencoded payload to the consumer", func() {
				for key, value := range validEncodedData {
					response, err := etcdClient.Set(key, value, 0)
					Expect(err).NotTo(HaveOccurred())

					Eventually(updates).Should(Receive(&response))
					Expect(response.PrevNode.Key).To(Equal(key))
					Expect(response.PrevNode.Value).To(Equal(validDecodedData[response.Node.Key]))
					Expect(response.Node.Key).To(Equal(key))
					Expect(response.Node.Value).To(Equal(validDecodedData[response.Node.Key]))
				}
			})
		})
	})
})

func basicSetup(etcdClient *etcdclient.Client) {
	for key, value := range validEncodedData {
		etcdClient.Set(key, value, 0)
	}

	for key, value := range invalidEncodedData {
		etcdClient.Set(key, value, 0)
	}
}

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
