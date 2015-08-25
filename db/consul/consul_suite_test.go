package consul_test

import (
	"github.com/cloudfoundry-incubator/bbs/db/consul"
	"github.com/cloudfoundry-incubator/bbs/db/consul/test/consul_helpers"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/consuladapter/consulrunner"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"testing"
)

var consulRunner *consulrunner.ClusterRunner
var consulSession *consuladapter.Session

var logger *lagertest.TestLogger
var consulHelper *consul_helpers.ConsulHelper

var consulDB *consul.ConsulDB

func TestDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul DB Suite")
}

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
	consulSession = consulRunner.NewSession("a-session")

	consulHelper = consul_helpers.NewConsulHelper(consulSession)
	consulDB = consul.NewConsul(consulSession)
})
