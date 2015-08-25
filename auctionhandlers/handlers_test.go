package auctionhandlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	fake_auction_runner "github.com/cloudfoundry-incubator/auction/auctiontypes/fakes"
	. "github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/rata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auction Handlers", func() {
	var (
		logger           *lagertest.TestLogger
		runner           *fake_auction_runner.FakeAuctionRunner
		responseRecorder *httptest.ResponseRecorder
		handler          http.Handler
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		runner = new(fake_auction_runner.FakeAuctionRunner)
		responseRecorder = httptest.NewRecorder()

		handler = New(runner, logger)
	})

	Describe("Task Handler", func() {
		Context("with a valid task", func() {
			BeforeEach(func() {
				tasks := []*models.Task{
					model_helpers.NewValidTask("the-task-guid"),
				}

				reqGen := rata.NewRequestGenerator("http://localhost", Routes)

				payload, err := json.Marshal(tasks)
				Expect(err).NotTo(HaveOccurred())

				req, err := reqGen.CreateRequest(CreateTaskAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				handler.ServeHTTP(responseRecorder, req)
			})

			It("responds with 202", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusAccepted))
			})

			It("logs with the correct session nesting", func() {
				Expect(logger.TestSink.LogMessages()).To(Equal([]string{
					"test.request.serving",
					"test.request.task-auction-handler.create.submitted",
					"test.request.done",
				}))

			})
		})
	})

	Describe("LRP Handler", func() {
		Context("with a valid LRPStart", func() {
			BeforeEach(func() {
				starts := []models.LRPStartRequest{{
					Indices: []uint{2},

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

				reqGen := rata.NewRequestGenerator("http://localhost", Routes)

				payload, err := json.Marshal(starts)
				Expect(err).NotTo(HaveOccurred())

				req, err := reqGen.CreateRequest(CreateLRPAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
				Expect(err).NotTo(HaveOccurred())

				handler.ServeHTTP(responseRecorder, req)
			})

			It("responds with 202", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusAccepted))
			})

			It("logs with the correct session nesting", func() {
				Expect(logger.TestSink.LogMessages()).To(Equal([]string{
					"test.request.serving",
					"test.request.lrp-auction-handler.create.submitted",
					"test.request.done",
				}))
			})
		})
	})
})
