package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Unavailable Handler", func() {
	var (
		fakeServer          *ghttp.Server
		handler             *handlers.UnavailableHandler
		serviceReady, ready chan struct{}

		responseRecorder *httptest.ResponseRecorder
		request          *http.Request
	)

	BeforeEach(func() {
		serviceReady = make(chan struct{})
		ready = make(chan struct{})

		fakeServer = ghttp.NewServer()
		handler = handlers.NewUnavailableHandler(fakeServer, serviceReady)
		responseRecorder = httptest.NewRecorder()

		var err error
		request, err = http.NewRequest("GET", "/test", nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("responds with 503 when the service is not ready", func() {
		handler.ServeHTTP(responseRecorder, request)
		Expect(responseRecorder.Code).To(Equal(http.StatusServiceUnavailable))
	})

	Context("when the service is ready", func() {
		BeforeEach(func() {
			fakeServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/test"),
				ghttp.RespondWith(200, nil, nil),
			))

			go func() {
				serviceReady <- struct{}{}
				ready <- struct{}{}
			}()
		})

		It("calls through to the wrapped handler", func() {
			Eventually(ready).Should(Receive())
			handler.ServeHTTP(responseRecorder, request)
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})
	})
})
