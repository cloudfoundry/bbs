package cellhandlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/cellhandlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PingHandler", func() {
	var (
		pingHandler *cellhandlers.PingHandler
		resp        *httptest.ResponseRecorder
		req         *http.Request
	)

	JustBeforeEach(func() {
		pingHandler = cellhandlers.NewPingHandler()

		resp = httptest.NewRecorder()

		var err error
		req, err = http.NewRequest("GET", "/ping", nil)
		Expect(err).NotTo(HaveOccurred())

		pingHandler.ServeHTTP(resp, req)
	})

	It("responds with 200 OK", func() {
		Expect(resp.Code).To(Equal(http.StatusOK))
	})
})
