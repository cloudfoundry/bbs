package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ETCD Metrics", func() {
	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	It("starts emitting metrics", func() {
		Eventually(testMetricsChan).Should(Receive())
	})
})
