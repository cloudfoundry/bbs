package trace_test

import (
	"bytes"
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/trace"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
)

var _ = Describe("RequestId", func() {
	var req *http.Request
	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("GET", "/info", bytes.NewReader([]byte("hello")))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("ContextWithRequestId", func() {
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

	Describe("RequestIdFromRequest", func() {
		It("returns empty request id from request", func() {
			Expect(trace.RequestIdFromRequest(req)).To(Equal(""))
		})

		It("returns request id from context", func() {
			req.Header.Add(trace.RequestIdHeader, "some-request-id")
			Expect(trace.RequestIdFromRequest(req)).To(Equal("some-request-id"))
		})
	})

	Describe("LoggerWithTraceInfo", func() {
		var logger lager.Logger
		var testSink *lagertest.TestSink

		BeforeEach(func() {
			logger = lager.NewLogger("test-logger")
			testSink = lagertest.NewTestSink()
			logger.RegisterSink(testSink)
		})

		Context("when trace id is empty", func() {
			It("does not set trace and span id", func() {
				logger = trace.LoggerWithTraceInfo(logger, "")
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data).To(BeEmpty())
				Expect(log.Data).To(BeEmpty())
			})
		})

		Context("when trace id is not empty", func() {
			It("sets trace and span id", func() {
				logger = trace.LoggerWithTraceInfo(logger, "7f461654-74d1-1ee5-8367-77d85df2cdab")
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log.Data["span-id"]).NotTo(BeEmpty())
			})

			It("generates new span id", func() {
				logger = trace.LoggerWithTraceInfo(logger, "7f461654-74d1-1ee5-8367-77d85df2cdab")
				logger.Info("test-log")

				log1 := testSink.Logs()[0]

				Expect(log1.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log1.Data["span-id"]).NotTo(BeEmpty())

				logger = trace.LoggerWithTraceInfo(logger, "7f461654-74d1-1ee5-8367-77d85df2cdab")
				logger.Info("test-log")

				log2 := testSink.Logs()[1]

				Expect(log2.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log2.Data["span-id"]).NotTo(BeEmpty())
				Expect(log2.Data["span-id"]).NotTo(Equal(log1.Data["span-id"]))
			})
		})

		Context("when trace id is invalid", func() {
			It("does not set trace and span id", func() {
				logger = trace.LoggerWithTraceInfo(logger, "invalid-request-id")
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data).To(BeEmpty())
				Expect(log.Data).To(BeEmpty())
			})
		})
	})
})
