package trace_test

import (
	"bytes"
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/trace"
)

var _ = Describe("RequestId", func() {
	Describe("ContextWithRequestId", func() {
		var req *http.Request
		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "/info", bytes.NewReader([]byte("hello")))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns context with no request id header", func() {
			ctx := trace.ContextWithRequestId(req)
			Expect(ctx.Value(trace.RequestIdHeader)).To(Equal(""))
		})

		It("returns context with request id header", func() {
			req.Header.Add(trace.RequestIdHeader, "some-request-id")
			ctx := trace.ContextWithRequestId(req)
			Expect(ctx.Value(trace.RequestIdHeader)).To(Equal("some-request-id"))
		})
	})

	Describe("RequestIdFromContext", func() {
		It("returns empty request id from context", func() {
			ctx := context.Background()
			Expect(trace.RequestIdFromContext(ctx)).To(Equal(""))
		})

		It("returns request id from context", func() {
			ctx := context.WithValue(context.Background(), trace.RequestIdHeader, "some-request-id")
			Expect(trace.RequestIdFromContext(ctx)).To(Equal("some-request-id"))
		})

	})
})
