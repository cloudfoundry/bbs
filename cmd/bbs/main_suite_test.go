package main_test

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/test_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
	"time"
)

var (
	etcdPort    int
	etcdUrl     string
	etcdRunner  *etcdstorerunner.ETCDClusterRunner
	etcdClient  *etcdclient.Client
	storeClient etcd.StoreClient

	logger lager.Logger

	client              bbs.InternalClient
	bbsBinPath          string
	bbsAddress          string
	bbsPort             int
	bbsURL              *url.URL
	bbsArgs             testrunner.Args
	bbsRunner           *ginkgomon.Runner
	bbsProcess          ifrit.Process
	consulRunner        *consulrunner.ClusterRunner
	consulClient        consuladapter.Client
	consulHelper        *test_helpers.ConsulHelper
	auctioneerServer    *ghttp.Server
	testMetricsListener net.PacketConn
	testMetricsChan     chan *events.Envelope

	sqlDBName string
	db        *sql.DB
)

var useSQL = flag.Bool("useSQL", false, "run integration tests against a local MySQL instance")

func TestBBS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Cmd Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		bbsConfig, err := gexec.Build("github.com/cloudfoundry-incubator/bbs/cmd/bbs", "-race")
		Expect(err).NotTo(HaveOccurred())
		return []byte(bbsConfig)
	},
	func(bbsConfig []byte) {
		bbsBinPath = string(bbsConfig)
		SetDefaultEventuallyTimeout(15 * time.Second)

		etcdPort = 4001 + GinkgoParallelNode()
		etcdUrl = fmt.Sprintf("http://127.0.0.1:%d", etcdPort)
		etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

		if *useSQL {
			// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
			// in diego-release.
			var err error
			db, err = sql.Open("mysql", "diego:diego_password@/")
			Expect(err).NotTo(HaveOccurred())
			Expect(db.Ping()).NotTo(HaveOccurred())

			_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
			Expect(err).NotTo(HaveOccurred())

			sqlDBName = fmt.Sprintf("diego_%d", GinkgoParallelNode())
			db, err = sql.Open("mysql", fmt.Sprintf("diego:diego_password@/%s", sqlDBName))
			Expect(err).NotTo(HaveOccurred())
			Expect(db.Ping()).NotTo(HaveOccurred())
		}

		consulRunner = consulrunner.NewClusterRunner(
			9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
			1,
			"http",
		)

		consulRunner.Start()
		consulRunner.WaitUntilReady()

		if !*useSQL {
			etcdRunner.Start()
		}
	},
)

var _ = SynchronizedAfterSuite(func() {
	if *useSQL {
		_, err := db.Exec(fmt.Sprintf("DROP DATABASE %s", sqlDBName))
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Close()).To(Succeed())
	}

	etcdRunner.Stop()
	consulRunner.Stop()
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	logger = lagertest.NewTestLogger("test")

	// etcdRunner.Reset()

	consulRunner.Reset()
	consulClient = consulRunner.NewClient()

	etcdClient = etcdRunner.Client()
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	auctioneerServer = ghttp.NewServer()
	auctioneerServer.UnhandledRequestStatusCode = http.StatusAccepted
	auctioneerServer.AllowUnhandledRequests = true

	bbsPort = 6700 + GinkgoParallelNode()
	bbsAddress = fmt.Sprintf("127.0.0.1:%d", bbsPort)

	bbsURL = &url.URL{
		Scheme: "http",
		Host:   bbsAddress,
	}

	testMetricsListener, _ = net.ListenPacket("udp", "127.0.0.1:0")
	testMetricsChan = make(chan *events.Envelope, 1)
	go func() {
		defer GinkgoRecover()
		for {
			buffer := make([]byte, 1024)
			n, _, err := testMetricsListener.ReadFrom(buffer)
			if err != nil {
				close(testMetricsChan)
				return
			}

			var envelope events.Envelope
			err = proto.Unmarshal(buffer[:n], &envelope)
			Expect(err).NotTo(HaveOccurred())
			testMetricsChan <- &envelope
		}
	}()

	port, err := strconv.Atoi(strings.TrimPrefix(testMetricsListener.LocalAddr().String(), "127.0.0.1:"))
	Expect(err).NotTo(HaveOccurred())

	client = bbs.NewClient(bbsURL.String())

	bbsArgs = testrunner.Args{
		Address:               bbsAddress,
		AdvertiseURL:          bbsURL.String(),
		AuctioneerAddress:     auctioneerServer.URL(),
		ConsulCluster:         consulRunner.ConsulCluster(),
		DropsondePort:         port,
		EtcdCluster:           etcdUrl,
		MetricsReportInterval: 10 * time.Millisecond,
		NoEtcd:                *useSQL,

		EncryptionKeys: []string{"label:key"},
		ActiveKeyLabel: "label",
	}
	if *useSQL {
		bbsArgs.DatabaseDriver = "mysql"
		bbsArgs.DatabaseConnectionString = fmt.Sprintf("diego:diego_password@/%s", sqlDBName)
	}
	storeClient = etcd.NewStoreClient(etcdClient)
	consulHelper = test_helpers.NewConsulHelper(logger, consulClient)
})

var _ = AfterEach(func() {
	if *useSQL {
		truncateTables(db)
	}
	auctioneerServer.Close()
	testMetricsListener.Close()
	Eventually(testMetricsChan).Should(BeClosed())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE domains",
	"TRUNCATE TABLE configurations",
	"TRUNCATE TABLE tasks",
	"TRUNCATE TABLE desired_lrps",
	"TRUNCATE TABLE actual_lrps",
}
