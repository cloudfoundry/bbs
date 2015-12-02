package encryptor_test

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var etcdPort int
var etcdUrl string
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var storeClient etcd.StoreClient

var etcdDB db.DB

func TestEncryptor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryptor Suite")
}

var _ = BeforeSuite(func() {
	etcdPort = 4001 + GinkgoParallelNode()
	etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

	etcdRunner.Start()
})

var _ = AfterSuite(func() {
	etcdRunner.Stop()
})

var _ = BeforeEach(func() {
	etcdRunner.Reset()

	etcdClient := etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	storeClient = etcd.NewStoreClient(etcdClient)
	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)
	etcdDB = etcd.NewETCD(format.ENCRYPTED_PROTO, 100, 100, 1*time.Minute, cryptor, storeClient, nil, nil, nil, nil, nil)
})
