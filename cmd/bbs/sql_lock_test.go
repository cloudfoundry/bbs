package main_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/locket"
	locketconfig "code.cloudfoundry.org/locket/cmd/locket/config"
	locketrunner "code.cloudfoundry.org/locket/cmd/locket/testrunner"
	"code.cloudfoundry.org/locket/lock"
	locketmodels "code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("SqlLock", func() {
	var (
		locketRunner  ifrit.Runner
		locketProcess ifrit.Process
		locketAddress string
	)

	BeforeEach(func() {
		locketPort, err := portAllocator.ClaimPorts(1)
		Expect(err).NotTo(HaveOccurred())

		locketAddress = fmt.Sprintf("localhost:%d", locketPort)
		locketRunner = locketrunner.NewLocketRunner(locketBinPath, func(cfg *locketconfig.LocketConfig) {
			cfg.DatabaseConnectionString = sqlRunner.ConnectionString()
			cfg.DatabaseDriver = sqlRunner.DriverName()
			cfg.ListenAddress = locketAddress
		})
		locketProcess = ginkgomon.Invoke(locketRunner)

		bbsConfig.ClientLocketConfig = locketrunner.ClientLocketConfig()
		bbsConfig.ClientLocketConfig.LocketAddress = locketAddress
	})

	JustBeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		// Give the BBS enough time to start
		bbsRunner.StartCheckTimeout = 4 * locket.RetryInterval
		bbsProcess = ifrit.Background(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Interrupt(bbsProcess)
		ginkgomon.Interrupt(locketProcess)
	})

	Context("with invalid configuration", func() {
		Context("when the locket address is not configured", func() {
			BeforeEach(func() {
				bbsConfig.LocketAddress = ""
			})

			It("exits with an error", func() {
				Eventually(bbsProcess.Wait()).Should(Receive(Not(BeNil())))
			})
		})

		Context("and the UUID is missing", func() {
			BeforeEach(func() {
				bbsConfig.UUID = ""
			})

			It("exits with an error", func() {
				Eventually(bbsProcess.Wait()).Should(Receive())
			})
		})
	})

	Context("with valid configuration", func() {
		JustBeforeEach(func() {
			Eventually(func() error {
				conn, err := net.Dial("tcp", bbsHealthAddress)
				if err != nil {
					return err
				}
				defer conn.Close()
				return nil
			}).ShouldNot(HaveOccurred())
		})

		It("acquires the lock in locket and becomes active", func() {
			Eventually(func() bool {
				return client.Ping(logger)
			}).Should(BeTrue())
		})

		It("has the configured UUID as the owner", func() {
			locketClient, err := locket.NewClient(logger, bbsConfig.ClientLocketConfig)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return client.Ping(logger)
			}).Should(BeTrue())

			lock, err := locketClient.Fetch(context.Background(), &locketmodels.FetchRequest{
				Key: "bbs",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(lock.Resource.Owner).To(Equal(bbsConfig.UUID))
		})

		It("emits metric about holding lock", func() {
			Eventually(func() bool {
				return client.Ping(logger)
			}).Should(BeTrue())

			Eventually(testMetricsChan).Should(Receive(
				testhelpers.MatchV2MetricAndValue(
					testhelpers.MetricAndValue{Name: "LockHeld", Value: 1},
				),
			))
		})

		Context("and the locking server becomes unreachable after grabbing the lock", func() {
			JustBeforeEach(func() {
				Eventually(func() bool {
					return client.Ping(logger)
				}).Should(BeTrue())

				ginkgomon.Interrupt(locketProcess)
			})

			It("exits", func() {
				// locket lock could take upto about 15 seconds to realize that the
				// lock is lost. add extra 2 seconds to give bbs a chance to exit
				Eventually(bbsProcess.Wait(), 17*time.Second).Should(Receive())
			})
		})

		Context("when locket enabled is true", func() {
			var competingProcess ifrit.Process

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
				competingRunner := lock.NewLockRunner(
					logger,
					locketClient,
					lockIdentifier,
					locket.DefaultSessionTTLInSeconds,
					clock,
					locket.RetryInterval,
				)
				competingProcess = ginkgomon.Invoke(competingRunner)
			})

			AfterEach(func() {
				ginkgomon.Interrupt(competingProcess)
			})

			It("blocks on waiting for the lock", func() {
				Consistently(func() bool {
					return client.Ping(logger)
				}).Should(BeFalse())
			})

			Context("and the lock becomes available", func() {
				JustBeforeEach(func() {
					Consistently(func() bool {
						return client.Ping(logger)
					}).Should(BeFalse())

					ginkgomon.Interrupt(competingProcess)
				})

				It("grabs the lock", func() {
					Eventually(func() bool {
						return client.Ping(logger)
					}, 5*locket.RetryInterval).Should(BeTrue())
				})
			})
		})
	})

	Context("when the sql lock is not available", func() {
		var competingProcess ifrit.Process

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
			competingRunner := lock.NewLockRunner(
				logger,
				locketClient,
				lockIdentifier,
				locket.DefaultSessionTTLInSeconds,
				clock,
				locket.RetryInterval,
			)
			competingProcess = ginkgomon.Invoke(competingRunner)
		})

		AfterEach(func() {
			ginkgomon.Interrupt(competingProcess)
		})

		It("does not become active", func() {
			Consistently(func() bool {
				return client.Ping(logger)
			}).Should(BeFalse())
		})

		It("emits metric about not holding lock", func() {
			Eventually(testMetricsChan).Should(Receive(
				testhelpers.MatchV2MetricAndValue(
					testhelpers.MetricAndValue{Name: "LockHeld", Value: 0},
				),
			))
		})

		Context("and continues to be unavailable", func() {
			It("exits", func() {
				Eventually(bbsProcess.Wait(), locket.DefaultSessionTTL*2).Should(Receive())
			})
		})

		Context("and the lock becomes available", func() {
			JustBeforeEach(func() {
				Consistently(func() bool {
					return client.Ping(logger)
				}).Should(BeFalse())

				ginkgomon.Interrupt(competingProcess)
			})

			It("grabs the lock and becomes active", func() {
				Eventually(func() bool {
					return client.Ping(logger)
				}, 5*locket.RetryInterval).Should(BeTrue())
			})
		})
	})
})
