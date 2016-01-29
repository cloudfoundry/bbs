package bbs_test

import (
	"github.com/cloudfoundry-incubator/bbs/test_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"testing"
)

var consulRunner *consulrunner.ClusterRunner
var consulClient consuladapter.Client

var logger *lagertest.TestLogger
var consulHelper *test_helpers.ConsulHelper

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("test")

	consulRunner = consulrunner.NewClusterRunner(
		9001+config.GinkgoConfig.ParallelNode*consulrunner.PortOffsetLength,
		1,
		"http",
	)

	logger = lagertest.NewTestLogger("test")

	consulRunner.Start()
	consulRunner.WaitUntilReady()
})

var _ = AfterSuite(func() {
	consulRunner.Stop()
})

var _ = BeforeEach(func() {
	consulRunner.Reset()
	consulClient = consulRunner.NewConsulClient()

	consulHelper = test_helpers.NewConsulHelper(logger, consulClient)
})

func TestBbs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Suite")
}
