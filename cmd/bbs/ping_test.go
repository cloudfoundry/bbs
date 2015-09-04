package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"
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
			BeforeEach(func() {
				err := consulSession.AcquireLock(shared.LockSchemaPath("bbs_lock"), []byte{})
				Expect(err).NotTo(HaveOccurred())

				bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
				bbsRunner.StartCheck = "bbs.lock-bbs.lock.acquiring-lock"

				bbsProcess = ginkgomon.Invoke(bbsRunner)
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
