package main_test

import (
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/lock"
	locketmodels "code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("Server Ready", func() {
	It("starts server that returns 503 even if lock is not acquired", func() {
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

		competingBBSLockProcess := ifrit.Invoke(competingBBSLockRunner)

		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsRunner.StartCheck = "bbs.locket-lock.started"
		bbsProcess = ginkgomon.Invoke(bbsRunner)

		By("returning 503 when bbs fails to acquire lock", func() {
			Consistently(func() bool {
				return client.Ping(logger, "some-trace-id")
			}).Should(BeFalse())
			_, err := client.Cells(logger, "some-trace-id")
			Expect(err).To(MatchError("Invalid Response with status code: 503"))
		})

		ginkgomon.Kill(competingBBSLockProcess)

		By("returning 200 when bbs acquires lock", func() {
			Eventually(func() bool {
				return client.Ping(logger, "some-trace-id")
			}).Should(BeTrue())
			_, err := client.Cells(logger, "some-trace-id")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
