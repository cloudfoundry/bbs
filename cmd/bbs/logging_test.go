package main_test

import (
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"

	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("enhanced BBS client debug logging", func() {
	Describe("a client creating a request", func() {
		It("includes the request name in the debug logs", func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsProcess = ginkgomon.Invoke(bbsRunner)

			_, err := client.Tasks(logger)
			Expect(err).NotTo(HaveOccurred())

			Eventually(logger).Should(gbytes.Say("request_name"))
			Eventually(logger).Should(gbytes.Say(bbs.TasksRoute_r3))
		})
	})

	Describe("a client submitting a request", func() {
		It("includes the request path in the debug logs", func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsProcess = ginkgomon.Invoke(bbsRunner)

			_, err := client.Tasks(logger)
			Expect(err).NotTo(HaveOccurred())

			Eventually(logger).Should(gbytes.Say("request_path"))
			Eventually(logger).Should(gbytes.Say(routePath(bbs.TasksRoute_r3)))
		})
	})

	Describe("a client completing a request", func() {
		It("includes the request path and duration in the debug logs", func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsProcess = ginkgomon.Invoke(bbsRunner)

			_, err := client.Tasks(logger)
			Expect(err).NotTo(HaveOccurred())

			Eventually(logger).Should(gbytes.Say("request_path"))
			Eventually(logger).Should(gbytes.Say(routePath(bbs.TasksRoute_r3)))
			Eventually(logger).Should(gbytes.Say("duration_in_ns"))
		})
	})

})

func routePath(reqName string) string {
	for _, route := range bbs.Routes {
		if route.Name == reqName {
			return route.Path
		}
	}

	return "no chance that this is a path"
}
