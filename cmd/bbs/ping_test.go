package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/pivotal-golang/clock"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping API", func() {
	Describe("Ping", func() {
		It("returns true when the bbs is running", func() {
			defer ginkgomon.Kill(bbsProcess)

			By("having the bbs down", func() {
				Expect(client.Ping(logger)).To(BeFalse())
			})

			By("starting the bbs without a lock", func() {
				competingBBSLock := locket.NewLock(logger, consulClient, locket.LockSchemaPath("bbs_lock"), []byte{}, clock.NewClock(), locket.RetryInterval, locket.LockTTL)
				competingBBSLockProcess := ifrit.Invoke(competingBBSLock)
				defer ginkgomon.Kill(competingBBSLockProcess)

				bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
				bbsRunner.StartCheck = "bbs.lock.acquiring-lock"
				bbsProcess = ginkgomon.Invoke(bbsRunner)

				Expect(client.Ping(logger)).To(BeFalse())
			})

			By("finally acquiring the lock", func() {
				Eventually(func() bool {
					return client.Ping(logger)
				}).Should(BeTrue())
			})
		})
	})
})
