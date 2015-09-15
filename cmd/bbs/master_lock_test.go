package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/shared"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MasterLock", func() {
	Context("when the bbs cannot obtain the bbs lock", func() {
		BeforeEach(func() {
			err := consulSession.AcquireLock(shared.LockSchemaPath("bbs_lock"), []byte{})
			Expect(err).NotTo(HaveOccurred())

			bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
			bbsRunner.StartCheck = "bbs.lock.acquiring-lock"

			bbsProcess = ginkgomon.Invoke(bbsRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(bbsProcess)
			consulSession.Destroy()
		})

		It("is not reachable", func() {
			_, err := client.Domains()
			Expect(err).To(HaveOccurred())
		})

		It("becomes available when the lock can be acquired", func() {
			consulSession.Destroy()
			_, err := consulSession.Recreate()
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				_, err := client.Domains()
				return err
			}).ShouldNot(HaveOccurred())
		})
	})

	Context("when the bbs loses the master lock", func() {
		BeforeEach(func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
			bbsProcess = ginkgomon.Invoke(bbsRunner)
			consulRunner.Reset()
		})

		AfterEach(func() {
			ginkgomon.Kill(bbsProcess)
		})

		It("exits with an error", func() {
			Eventually(bbsRunner.ExitCode, 3).Should(Equal(1))
		})
	})
})
