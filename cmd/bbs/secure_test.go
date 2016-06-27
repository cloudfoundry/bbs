package main_test

import (
	"os"
	"path"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Secure", func() {
	var (
		client bbs.InternalClient
		err    error

		basePath string
	)

	BeforeEach(func() {
		basePath = path.Join(os.Getenv("GOPATH"), "src/code.cloudfoundry.org/bbs/cmd/bbs/fixtures")
		bbsURL.Scheme = "https"
	})

	JustBeforeEach(func() {
		client = bbs.NewClient(bbsURL.String())
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	Context("when configuring mutual SSL", func() {
		BeforeEach(func() {
			bbsArgs.RequireSSL = true
			bbsArgs.CAFile = path.Join(basePath, "green-certs", "server-ca.crt")
			bbsArgs.CertFile = path.Join(basePath, "green-certs", "server.crt")
			bbsArgs.KeyFile = path.Join(basePath, "green-certs", "server.key")
		})

		It("succeeds for a client configured with the right certificate", func() {
			caFile := path.Join(basePath, "green-certs", "server-ca.crt")
			certFile := path.Join(basePath, "green-certs", "client.crt")
			keyFile := path.Join(basePath, "green-certs", "client.key")
			client, err = bbs.NewSecureClient(bbsURL.String(), caFile, certFile, keyFile, 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.Ping(logger)).To(BeTrue())
		})

		It("fails for a client with no SSL", func() {
			client = bbs.NewClient(bbsURL.String())
			Expect(client.Ping(logger)).To(BeFalse())
		})

		It("fails for a client configured with the wrong certificates", func() {
			caFile := path.Join(basePath, "green-certs", "server-ca.crt")
			certFile := path.Join(basePath, "blue-certs", "client.crt")
			keyFile := path.Join(basePath, "blue-certs", "client.key")
			client, err = bbs.NewSecureClient(bbsURL.String(), caFile, certFile, keyFile, 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.Ping(logger)).To(BeFalse())
		})

		It("fails for a client configured with a different ca certificate", func() {
			caFile := path.Join(basePath, "blue-certs", "server-ca.crt")
			certFile := path.Join(basePath, "green-certs", "client.crt")
			keyFile := path.Join(basePath, "green-certs", "client.key")
			client, err = bbs.NewSecureClient(bbsURL.String(), caFile, certFile, keyFile, 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.Ping(logger)).To(BeFalse())
		})

		It("fails to create the client if certs are not valid", func() {
			client, err = bbs.NewSecureClient(bbsURL.String(), "", "", "", 0, 0)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when configuring a client without mutual SSL (skipping verification)", func() {
		BeforeEach(func() {
			bbsArgs.RequireSSL = true
			bbsArgs.CertFile = path.Join(basePath, "green-certs", "server.crt")
			bbsArgs.KeyFile = path.Join(basePath, "green-certs", "server.key")
		})

		It("succeeds for a client configured with the right certificate", func() {
			certFile := path.Join(basePath, "green-certs", "client.crt")
			keyFile := path.Join(basePath, "green-certs", "client.key")
			client, err = bbs.NewSecureSkipVerifyClient(bbsURL.String(), certFile, keyFile, 0, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails for a client configured with the wrong certificates", func() {
			certFile := path.Join(basePath, "blue-certs", "client.crt")
			keyFile := path.Join(basePath, "blue-certs", "client.key")
			client, err = bbs.NewSecureSkipVerifyClient(bbsURL.String(), certFile, keyFile, 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.Ping(logger)).To(BeFalse())
		})
	})
})
