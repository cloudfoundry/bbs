package measurements_test

import (
	"os"
	"path"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS", func() {
	var basePath string

	BeforeEach(func() {
		basePath = path.Join(os.Getenv("GOPATH"), "src", "github.com", "cloudfoundry-incubator", "bbs", "cmd", "bbs", "fixtures")
	})

	desireLRP := func() {
		var err error
		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())
		desiredLRP := model_helpers.NewValidDesiredLRP(guid.String())
		desiredLRP.Instances = 8
		err = client.DesireLRP(desiredLRP)
		Expect(err).NotTo(HaveOccurred())
		readLRP, err := client.DesiredLRPByProcessGuid(guid.String())
		Expect(err).NotTo(HaveOccurred())
		Expect(readLRP).ToNot(BeNil())
	}

	runMeasurements := func() {
		Measure("ping time", func(b Benchmarker) {
			b.Time("first-request", func() {
				Expect(client.Ping()).To(BeTrue())
			})
			b.Time("second-request", func() {
				Expect(client.Ping()).To(BeTrue())
			})
		}, 3)

		Measure("desire lrp time", func(b Benchmarker) {
			b.Time("first-request", func() {
				desireLRP()
			})
			b.Time("second-request", func() {
				desireLRP()
			})
		}, 3)
	}

	Context("when configuring mutual SSL", func() {
		BeforeEach(func() {
			etcdSSLConfig = &etcdstorerunner.SSLConfig{
				CAFile:   path.Join(basePath, "blue-certs", "server-ca.crt"),
				CertFile: path.Join(basePath, "blue-certs", "server.crt"),
				KeyFile:  path.Join(basePath, "blue-certs", "server.key"),
			}

			bbsArgs.EtcdCACert = path.Join(basePath, "blue-certs", "server-ca.crt")
			bbsArgs.EtcdClientCert = path.Join(basePath, "blue-certs", "client.crt")
			bbsArgs.EtcdClientKey = path.Join(basePath, "blue-certs", "client.key")

			bbsURL.Scheme = "https"

			bbsArgs.RequireSSL = true
			bbsArgs.CAFile = path.Join(basePath, "green-certs", "server-ca.crt")
			bbsArgs.CertFile = path.Join(basePath, "green-certs", "server.crt")
			bbsArgs.KeyFile = path.Join(basePath, "green-certs", "server.key")

			caFile := path.Join(basePath, "green-certs", "server-ca.crt")
			certFile := path.Join(basePath, "green-certs", "client.crt")
			keyFile := path.Join(basePath, "green-certs", "client.key")

			var err error
			client, err = bbs.NewSecureClient(bbsURL.String(), caFile, certFile, keyFile)
			Expect(err).NotTo(HaveOccurred())
		})

		runMeasurements()
	})

	Context("when NOT configuring mutual SSL", func() {
		BeforeEach(func() {
			etcdSSLConfig = nil
			bbsURL.Scheme = "http"
			bbsArgs.RequireSSL = false
			client = bbs.NewClient(bbsURL.String())
		})

		runMeasurements()
	})
})
