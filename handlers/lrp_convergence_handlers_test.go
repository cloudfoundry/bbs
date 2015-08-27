package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("LRP Convergence Handlers", func() {
	var (
		logger           lager.Logger
		fakeLRPDB        *fakes.FakeLRPDB
		responseRecorder *httptest.ResponseRecorder

		handler *handlers.LRPConvergenceHandler
	)

	BeforeEach(func() {
		fakeLRPDB = new(fakes.FakeLRPDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewLRPConvergenceHandler(logger, fakeLRPDB)
	})

	JustBeforeEach(func() {
		handler.ConvergeLRPs(responseRecorder, nil)
	})

	It("calls ConvergeLRPs", func() {
		Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		Expect(fakeLRPDB.ConvergeLRPsCallCount()).To(Equal(1))
	})

	It("returns a Ok response", func() {
		Expect(responseRecorder.Code).To(Equal(http.StatusOK))
	})
})
