package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/lock"
	locketmodels "code.cloudfoundry.org/locket/models"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MasterLock", func() {
	Context("when the bbs cannot obtain the bbs lock", func() {
		var competingBBSLockProcess ifrit.Process

		BeforeEach(func() {
			locketClient, err := locket.NewClient(logger, bbsConfig.ClientLocketConfig)
			Expect(err).NotTo(HaveOccurred())

			lockIdentifier := &locketmodels.Resource{
				Key:      "bbs",
				Owner:    "Your worst enemy.",
				Value:    "Something",
				TypeCode: locketmodels.LOCK,
			}

			clock := clock.NewClock()
			competingBBSLockRunner := lock.NewLockRunner(
				logger,
				locketClient,
				lockIdentifier,
				locket.DefaultSessionTTLInSeconds,
				clock,
				locket.RetryInterval,
			)
			competingBBSLockProcess = ginkgomon.Invoke(competingBBSLockRunner)

			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsRunner.StartCheck = "bbs.locket-lock.started"

			bbsProcess = ginkgomon.Invoke(bbsRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(competingBBSLockProcess)
		})

		It("is not reachable", func() {
			_, err := client.Domains(logger)
			Expect(err).To(HaveOccurred())
		})

		It("becomes available when the lock can be acquired", func() {
			ginkgomon.Kill(competingBBSLockProcess)

			Eventually(func() error {
				_, err := client.Domains(logger)
				return err
			}).ShouldNot(HaveOccurred())
		})
	})

	Context("when the bbs loses the master lock", func() {
		BeforeEach(func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsProcess = ginkgomon.Invoke(bbsRunner)
			ginkgomon.Kill(locketProcess)
		})

		It("exits with an error", func() {
			Eventually(bbsRunner.ExitCode, 16).Should(Equal(1))
		})
	})
})
