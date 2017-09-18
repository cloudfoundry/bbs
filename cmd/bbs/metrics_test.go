package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	sonde_events "github.com/cloudfoundry/sonde-go/events"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics", func() {
	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	It("starts emitting metrics", func() {
		Eventually(testMetricsChan).Should(Receive())
	})

	It("starts emitting file-descriptor count metrics", func() {
		Eventually(func() string {
			metric := <-testMetricsChan
			if metric.GetEventType() == sonde_events.Envelope_ValueMetric {
				return *metric.ValueMetric.Name
			}
			return ""
		}).Should(Equal("OpenFileDescriptors"))
	})
})
