package bbs_test

import (
	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/consuladapter/consulrunner"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

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
	_ = consulRunner.Reset()
	consulClient = consulRunner.NewClient()

	consulHelper = test_helpers.NewConsulHelper(logger, consulClient)
})

func TestBbs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BBS Suite")
}
