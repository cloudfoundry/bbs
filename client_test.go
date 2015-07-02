package bbs_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var (
	registryHost  string
	fakeBBSServer *ghttp.Server
	client        bbs.Client
)

var _ = Describe("BBS Client", func() {
	BeforeEach(func() {
		fakeBBSServer = ghttp.NewServer()
		client = bbs.NewClient(fakeBBSServer.URL())
	})

	AfterEach(func() {
		fakeBBSServer.Close()
	})

	Describe("Client Request", func() {
		var domains []string

		BeforeEach(func() {
			domains = []string{"some-domain"}

			fakeBBSServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/domains"),
				ghttp.VerifyContentType(bbs.ProtoContentType),
				ghttp.RespondWithJSONEncoded(http.StatusOK, domains),
			))
		})

		It("sends a valid request from the client to the bbs", func() {
			response, err := client.Domains()

			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(domains))
		})
	})
})
