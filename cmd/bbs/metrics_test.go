package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ETCD Metrics", func() {
	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	It("starts emitting metrics", func() {
		Eventually(testMetricsChan).Should(Receive())
	})
})
