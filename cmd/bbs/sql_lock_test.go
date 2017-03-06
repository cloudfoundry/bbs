package main_test

import (
	"crypto/tls"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/localip"
	"code.cloudfoundry.org/locket"
	locketconfig "code.cloudfoundry.org/locket/cmd/locket/config"
	locketrunner "code.cloudfoundry.org/locket/cmd/locket/testrunner"
	"code.cloudfoundry.org/locket/lock"
	locketmodels "code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("SqlLock", func() {
	var (
		locketRunner    ifrit.Runner
		locketProcess   ifrit.Process
		locketAddress   string
		locketTLSConfig *tls.Config
	)

	BeforeEach(func() {
		locketPort, err := localip.LocalPort()
		Expect(err).NotTo(HaveOccurred())

		caFile := "fixtures/locket-certs/ca.crt"
		certFile := "fixtures/locket-certs/cert.crt"
		keyFile := "fixtures/locket-certs/key.key"

		locketAddress = fmt.Sprintf("localhost:%d", locketPort)
		locketConfig := locketconfig.LocketConfig{
			CaFile:                   caFile,
			CertFile:                 certFile,
			ConsulCluster:            consulRunner.ConsulCluster(),
			DatabaseConnectionString: sqlRunner.ConnectionString(),
			DatabaseDriver:           sqlRunner.DriverName(),
			KeyFile:                  keyFile,
			ListenAddress:            locketAddress,
		}

		locketRunner = locketrunner.NewLocketRunner(locketBinPath, locketConfig)
		locketProcess = ginkgomon.Invoke(locketRunner)

		bbsConfig.LocketAddress = locketAddress
		bbsConfig.LocketCACert = caFile
		bbsConfig.LocketClientCert = certFile
		bbsConfig.LocketClientKey = keyFile

		locketTLSConfig, err = cfhttp.NewTLSConfig(certFile, keyFile, caFile)
		Expect(err).NotTo(HaveOccurred())
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
		Context("and the locket address is not configured", func() {
			BeforeEach(func() {
				bbsConfig.LocketAddress = ""
				bbsConfig.SkipConsulLock = true
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

		Context("and the locking server becomes unreachable after grabbing the lock", func() {
			JustBeforeEach(func() {
				Eventually(func() bool {
					return client.Ping(logger)
				}).Should(BeTrue())

				ginkgomon.Interrupt(locketProcess)
			})

			It("exits", func() {
				Eventually(bbsProcess.Wait()).Should(Receive())
			})
		})

		Context("when consul lock isn't required", func() {
			var competingBBSLockProcess ifrit.Process

			BeforeEach(func() {
				bbsConfig.SkipConsulLock = true
				competingBBSLock := locket.NewLock(logger, consulClient, locket.LockSchemaPath("bbs_lock"), []byte{}, clock.NewClock(), locket.RetryInterval, locket.DefaultSessionTTL)
				competingBBSLockProcess = ifrit.Invoke(competingBBSLock)
			})

			AfterEach(func() {
				ginkgomon.Kill(competingBBSLockProcess)
			})

			It("does not acquire the consul lock", func() {
				Eventually(func() bool {
					return client.Ping(logger)
				}).Should(BeTrue())
			})
		})

		Context("when the lock is not available", func() {
			var competingProcess ifrit.Process

			BeforeEach(func() {
				conn, err := grpc.Dial(locketAddress, grpc.WithTransportCredentials(credentials.NewTLS(locketTLSConfig)))
				Expect(err).NotTo(HaveOccurred())
				locketClient := locketmodels.NewLocketClient(conn)

				lockIdentifier := &locketmodels.Resource{
					Key:   "bbs",
					Owner: "Your worst enemy.",
					Value: "Something",
				}

				clock := clock.NewClock()
				competingRunner := lock.NewLockRunner(logger, locketClient, lockIdentifier, 5, clock, locket.RetryInterval)
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
})
