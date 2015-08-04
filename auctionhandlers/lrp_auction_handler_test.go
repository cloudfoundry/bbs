package auctionhandlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	fake_auction_runner "github.com/cloudfoundry-incubator/auction/auctiontypes/fakes"
	"github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LRPAuctionHandler", func() {
	var (
		logger           lager.Logger
		runner           *fake_auction_runner.FakeAuctionRunner
		responseRecorder *httptest.ResponseRecorder
		handler          *auctionhandlers.LRPAuctionHandler
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		runner = new(fake_auction_runner.FakeAuctionRunner)
		responseRecorder = httptest.NewRecorder()
		handler = auctionhandlers.NewLRPAuctionHandler(runner)
	})

	Describe("Create", func() {
		Context("when the request body is an LRP start auction request", func() {
			var starts []models.LRPStartRequest

			BeforeEach(func() {
				starts = []models.LRPStartRequest{{
					Indices: []uint{2, 3},

					DesiredLRP: &models.DesiredLRP{
						Domain:      "tests",
						ProcessGuid: "some-guid",

						RootFs:    "docker:///docker.com/docker",
						Instances: 1,
						MemoryMb:  1024,
						DiskMb:    512,
						CpuWeight: 42,
						Action: models.WrapAction(&models.DownloadAction{
							From: "http://example.com",
							To:   "/tmp/internet",
							User: "diego",
						}),
					},
				}}

				handler.Create(responseRecorder, newTestRequest(starts), logger)
			})

			It("responds with 202", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusAccepted))
			})

			It("responds with an empty JSON body", func() {
				Expect(responseRecorder.Body.String()).To(Equal("{}"))
			})

			It("should submit the start auction to the auction runner", func() {
				Expect(runner.ScheduleLRPsForAuctionsCallCount()).To(Equal(1))

				submittedStart := runner.ScheduleLRPsForAuctionsArgsForCall(0)
				Expect(submittedStart).To(Equal(starts))
			})
		})

		Context("when the start auction has invalid index", func() {
			var start models.LRPStartRequest

			BeforeEach(func() {
				start = models.LRPStartRequest{}

				handler.Create(responseRecorder, newTestRequest(start), logger)
			})

			It("responds with 400", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a JSON body containing the error", func() {
				handlerError := auctionhandlers.HandlerError{}
				err := json.NewDecoder(responseRecorder.Body).Decode(&handlerError)
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerError.Error).NotTo(BeEmpty())
			})

			It("should not submit the start auction to the auction runner", func() {
				Expect(runner.ScheduleLRPsForAuctionsCallCount()).To(Equal(0))
			})
		})

		Context("when the request body is a not a start auction", func() {
			BeforeEach(func() {
				handler.Create(responseRecorder, newTestRequest(`{invalidjson}`), logger)
			})

			It("responds with 400", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("responds with a JSON body containing the error", func() {
				handlerError := auctionhandlers.HandlerError{}
				err := json.NewDecoder(responseRecorder.Body).Decode(&handlerError)
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerError.Error).NotTo(BeEmpty())
			})

			It("should not submit the start auction to the auction runner", func() {
				Expect(runner.ScheduleLRPsForAuctionsCallCount()).To(Equal(0))
			})
		})

		Context("when the request body returns a non-EOF error on read", func() {
			BeforeEach(func() {
				req := newTestRequest("")
				req.Body = badReader{}
				handler.Create(responseRecorder, req, logger)
			})

			It("responds with 500", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("responds with a JSON body containing the error", func() {
				handlerError := auctionhandlers.HandlerError{}
				err := json.NewDecoder(responseRecorder.Body).Decode(&handlerError)
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerError.Error).To(Equal(ErrBadRead.Error()))
			})

			It("should not submit the start auction to the auction runner", func() {
				Expect(runner.ScheduleLRPsForAuctionsCallCount()).To(Equal(0))
			})
		})
	})
})
