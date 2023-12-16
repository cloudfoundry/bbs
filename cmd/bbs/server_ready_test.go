package main_test

import (
	"net/http"

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

var _ = FDescribe("Server Ready", func() {
	pingStatusCode := func() int {
		resp, err := http.Get("http://" + bbsAddress + "/v1/events")
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

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
		ginkgomon.Invoke(bbsRunner)

		By("bbs failing to acquire lock due to competing lock", func() {
			Consistently(pingStatusCode).Should(Equal(http.StatusServiceUnavailable))
		})

		ginkgomon.Kill(competingBBSLockProcess)

		By("returning 200 when bbs acquires lock", func() {
			Eventually(pingStatusCode).Should(Equal(http.StatusOK))
		})
	})
})
