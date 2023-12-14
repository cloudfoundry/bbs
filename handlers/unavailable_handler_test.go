package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/bbs/handlers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Unavailable Handler", func() {
	var (
		fakeServer   *ghttp.Server
		handler      *handlers.UnavailableHandler
		serviceReady chan struct{}

		request *http.Request
	)

	BeforeEach(func() {
		serviceReady = make(chan struct{})

		fakeServer = ghttp.NewServer()
		handler = handlers.NewUnavailableHandler(fakeServer, serviceReady)

		var err error
		request, err = http.NewRequest("GET", "/test", nil)
		Expect(err).NotTo(HaveOccurred())

		fakeServer.RouteToHandler("GET", "/test", ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/test"),
			ghttp.RespondWith(200, nil, nil),
		))
	})

	verifyEventualResponse := func(expectedStatus int, handler *handlers.UnavailableHandler) {
		EventuallyWithOffset(1, func() int {
			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, request)
			return responseRecorder.Code
		}).Should(Equal(expectedStatus))
	}

	verifyConsistentResponse := func(expectedStatus int, handler *handlers.UnavailableHandler) {
		ConsistentlyWithOffset(1, func() int {
			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, request)
			return responseRecorder.Code
		}).Should(Equal(expectedStatus))
	}

	It("responds with 503 until the service is ready", func() {
		verifyConsistentResponse(http.StatusServiceUnavailable, handler)

		close(serviceReady)

		verifyEventualResponse(http.StatusOK, handler)
		verifyConsistentResponse(http.StatusOK, handler)
	})

	Context("when there are multiple channels specifying whether the service is ready", func() {
		var serviceReady2 chan struct{}

		BeforeEach(func() {
			serviceReady2 = make(chan struct{})
			handler = handlers.NewUnavailableHandler(fakeServer, serviceReady, serviceReady2)
		})

		It("responds with 503 until both channels have been closed", func() {
			verifyConsistentResponse(http.StatusServiceUnavailable, handler)

			close(serviceReady)

			verifyConsistentResponse(http.StatusServiceUnavailable, handler)

			close(serviceReady2)

			verifyEventualResponse(http.StatusOK, handler)
			verifyConsistentResponse(http.StatusOK, handler)
		})

	})
})
