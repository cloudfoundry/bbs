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
	AfterEach(func() {
		if bbsProcess != nil {
			ginkgomon.Kill(bbsProcess)
		}
	})

	Describe("Ping", func() {
		Context("when the BBS Server is up", func() {
			BeforeEach(func() {
				bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
				bbsProcess = ginkgomon.Invoke(bbsRunner)
			})

			It("returns true", func() {
				Expect(client.Ping()).To(BeTrue())
			})
		})

		Context("when the BBS Server is not the leader", func() {
			var competingBBSLockProcess ifrit.Process
			BeforeEach(func() {
				competingBBSLock := locket.NewLock(logger, consulClient, locket.LockSchemaPath("bbs_lock"), []byte{}, clock.NewClock(), locket.RetryInterval, locket.LockTTL)
				competingBBSLockProcess = ifrit.Invoke(competingBBSLock)

				bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
				bbsRunner.StartCheck = "bbs.lock.acquiring-lock"

				bbsProcess = ginkgomon.Invoke(bbsRunner)
			})

			AfterEach(func() {
				ginkgomon.Kill(competingBBSLockProcess)
			})

			It("returns false", func() {
				Expect(client.Ping()).To(BeFalse())
			})
		})

		Context("when the BBS Server down", func() {
			It("returns false", func() {
				Expect(client.Ping()).To(BeFalse())
			})
		})
	})
})
