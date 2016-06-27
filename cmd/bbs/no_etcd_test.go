package main_test

import (
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("BBS With Only SQL", func() {
	if test_helpers.UseSQL() {
		BeforeEach(func() {
			port, err := strconv.Atoi(strings.TrimPrefix(testMetricsListener.LocalAddr().String(), "127.0.0.1:"))
			Expect(err).NotTo(HaveOccurred())

			bbsArgs = testrunner.Args{
				Address:               bbsAddress,
				HealthAddress:         bbsHealthAddress,
				AdvertiseURL:          bbsURL.String(),
				AuctioneerAddress:     auctioneerServer.URL(),
				ConsulCluster:         consulRunner.ConsulCluster(),
				DropsondePort:         port,
				MetricsReportInterval: 10 * time.Millisecond,
				EncryptionKeys:        []string{"label:key"},
				ActiveKeyLabel:        "label",
			}
		})

		JustBeforeEach(func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		})

		Context("when etcd is partially configured", func() {
			BeforeEach(func() {
				bbsArgs.EtcdCACert = "I am a ca cert"
			})

			It("returns a validation error", func() {
				bbsProcess = ifrit.Invoke(bbsRunner)
				Eventually(bbsProcess.Wait()).Should(Receive(HaveOccurred()))
			})
		})

		Context("when etcd is not configured at all", func() {
			Context("and sql is configured", func() {
				BeforeEach(func() {
					bbsArgs.DatabaseDriver = sqlRunner.DriverName()
					bbsArgs.DatabaseConnectionString = sqlRunner.ConnectionString()
				})

				It("the bbs eventually responds to ping", func() {
					bbsProcess = ginkgomon.Invoke(bbsRunner)
					Expect(client.Ping(logger)).To(BeTrue())
				})
			})

			Context("when sql is not configured", func() {
				It("the bbs returns a validation error", func() {
					bbsProcess = ifrit.Invoke(bbsRunner)
					Eventually(bbsProcess.Wait()).Should(Receive(HaveOccurred()))
				})
			})
		})
	}
})
