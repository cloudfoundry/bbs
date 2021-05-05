package rep_test

import (
	"testing"
	"time"

	"code.cloudfoundry.org/bbs/clients/rep"
	"code.cloudfoundry.org/bbs/clients/rep/repfakes"
	cfhttp "code.cloudfoundry.org/cfhttp/v2"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/consuladapter/consulrunner"
	executorfakes "code.cloudfoundry.org/executor/fakes"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var (
	cfHttpTimeout time.Duration
	auctionRep    *repfakes.FakeClient
	factory       rep.ClientFactory

	fakeExecutorClient *executorfakes.FakeClient
	consulRunner       *consulrunner.ClusterRunner
	consulClient       consuladapter.Client
)

func TestRep(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rep Suite")
}

var _ = BeforeSuite(func() {
	consulRunner = consulrunner.NewClusterRunner(
		consulrunner.ClusterRunnerConfig{
			StartingPort: 9001 + config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
			NumNodes:     1,
			Scheme:       "http",
		},
	)

	consulRunner.Start()
	consulRunner.WaitUntilReady()
})

var _ = AfterSuite(func() {
	consulRunner.Stop()
})

var _ = BeforeEach(func() {
	auctionRep = &repfakes.FakeClient{}
	fakeExecutorClient = &executorfakes.FakeClient{}
	consulRunner.Reset()
	consulClient = consulRunner.NewClient()

	cfHttpTimeout = 1 * time.Second
	var err error
	httpClient := cfhttp.NewClient(
		cfhttp.WithRequestTimeout(cfHttpTimeout),
	)
	factory, err = rep.NewClientFactory(httpClient, httpClient, nil)
	Expect(err).NotTo(HaveOccurred())
})
