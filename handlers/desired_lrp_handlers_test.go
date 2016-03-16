package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/auctioneer/auctioneerfakes"
	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	"github.com/cloudfoundry-incubator/rep"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("DesiredLRP Handlers", func() {
	var (
		logger               *lagertest.TestLogger
		fakeDesiredLRPDB     *fakes.FakeDesiredLRPDB
		fakeActualLRPDB      *fakes.FakeActualLRPDB
		fakeAuctioneerClient *auctioneerfakes.FakeClient

		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.DesiredLRPHandler

		desiredLRP1 models.DesiredLRP
		desiredLRP2 models.DesiredLRP
	)

	BeforeEach(func() {
		fakeDesiredLRPDB = new(fakes.FakeDesiredLRPDB)
		fakeActualLRPDB = new(fakes.FakeActualLRPDB)
		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
		logger = lagertest.NewTestLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewDesiredLRPHandler(logger, 5, fakeDesiredLRPDB, fakeActualLRPDB, fakeAuctioneerClient, fakeRepClientFactory, fakeServiceClient)
	})

	Describe("DesiredLRPs", func() {
		var requestBody interface{}

		BeforeEach(func() {
			requestBody = &models.DesiredLRPsRequest{}
			desiredLRP1 = models.DesiredLRP{}
			desiredLRP2 = models.DesiredLRP{}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesiredLRPs(responseRecorder, request)
		})

		Context("when reading desired lrps from DB succeeds", func() {
			var desiredLRPs []*models.DesiredLRP

			BeforeEach(func() {
				desiredLRPs = []*models.DesiredLRP{&desiredLRP1, &desiredLRP2}
				fakeDesiredLRPDB.DesiredLRPsReturns(desiredLRPs, nil)
			})

			It("returns a list of desired lrp groups", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrps).To(Equal(desiredLRPs))
			})

			Context("and no filter is provided", func() {
				It("call the DB with no filters to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter).To(Equal(models.DesiredLRPFilter{}))
				})
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					requestBody = &models.DesiredLRPsRequest{Domain: "domain-1"}
				})

				It("call the DB with the domain filter to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPsCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPsArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})
		})

		Context("when the DB returns no desired lrp groups", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns([]*models.DesiredLRP{}, nil)
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrps).To(BeEmpty())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPsReturns([]*models.DesiredLRP{}, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesiredLRPByProcessGuid", func() {
		var (
			processGuid = "process-guid"

			requestBody interface{}
		)

		BeforeEach(func() {
			requestBody = &models.DesiredLRPByProcessGuidRequest{
				ProcessGuid: processGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesiredLRPByProcessGuid(responseRecorder, request)
		})

		Context("when reading desired lrp from DB succeeds", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = &models.DesiredLRP{ProcessGuid: processGuid}
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
			})

			It("fetches desired lrp by process guid", func() {
				Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
				_, actualProcessGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrp).To(Equal(desiredLRP))
			})
		})

		Context("when the DB returns no desired lrp", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, models.ErrResourceNotFound)
			})

			It("returns a resource not found error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesiredLRPSchedulingInfos", func() {
		var (
			requestBody     interface{}
			schedulingInfo1 models.DesiredLRPSchedulingInfo
			schedulingInfo2 models.DesiredLRPSchedulingInfo
		)

		BeforeEach(func() {
			requestBody = &models.DesiredLRPsRequest{}
			schedulingInfo1 = models.DesiredLRPSchedulingInfo{}
			schedulingInfo2 = models.DesiredLRPSchedulingInfo{}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesiredLRPSchedulingInfos(responseRecorder, request)
		})

		Context("when reading scheduling infos from DB succeeds", func() {
			var schedulingInfos []*models.DesiredLRPSchedulingInfo

			BeforeEach(func() {
				schedulingInfos = []*models.DesiredLRPSchedulingInfo{&schedulingInfo1, &schedulingInfo2}
				fakeDesiredLRPDB.DesiredLRPSchedulingInfosReturns(schedulingInfos, nil)
			})

			It("returns a list of desired lrp groups", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPSchedulingInfosResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrpSchedulingInfos).To(Equal(schedulingInfos))
			})

			Context("and no filter is provided", func() {
				It("call the DB with no filters to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPSchedulingInfosCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPSchedulingInfosArgsForCall(0)
					Expect(filter).To(Equal(models.DesiredLRPFilter{}))
				})
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					requestBody = &models.DesiredLRPsRequest{Domain: "domain-1"}
				})

				It("call the DB with the domain filter to retrieve the desired lrps", func() {
					Expect(fakeDesiredLRPDB.DesiredLRPSchedulingInfosCallCount()).To(Equal(1))
					_, filter := fakeDesiredLRPDB.DesiredLRPSchedulingInfosArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})
		})

		Context("when the DB returns no desired lrp groups", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPSchedulingInfosReturns([]*models.DesiredLRPSchedulingInfo{}, nil)
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPSchedulingInfosResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.DesiredLrpSchedulingInfos).To(BeEmpty())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesiredLRPSchedulingInfosReturns([]*models.DesiredLRPSchedulingInfo{}, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPSchedulingInfosResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesireDesiredLRP", func() {
		var (
			desiredLRP *models.DesiredLRP

			requestBody interface{}
		)

		BeforeEach(func() {
			desiredLRP = model_helpers.NewValidDesiredLRP("some-guid")
			desiredLRP.Instances = 5
			requestBody = &models.DesireLRPRequest{
				DesiredLrp: desiredLRP,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesireDesiredLRP(responseRecorder, request)
		})

		Context("when creating desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesireLRPReturns(nil)
				fakeActualLRPDB.CreateUnclaimedActualLRPReturns(nil)
			})

			It("creates desired lrp", func() {
				Expect(fakeDesiredLRPDB.DesireLRPCallCount()).To(Equal(1))
				_, actualDesiredLRP := fakeDesiredLRPDB.DesireLRPArgsForCall(0)
				Expect(actualDesiredLRP).To(Equal(desiredLRP))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			It("creates one ActualLRP per index", func() {
				Expect(fakeActualLRPDB.CreateUnclaimedActualLRPCallCount()).To(Equal(5))

				for i := 0; i < 5; i++ {
					_, actualLRPKey := fakeActualLRPDB.CreateUnclaimedActualLRPArgsForCall(i)
					expectedLRPKey := &models.ActualLRPKey{
						ProcessGuid: desiredLRP.ProcessGuid,
						Domain:      desiredLRP.Domain,
						Index:       int32(i),
					}
					Expect(actualLRPKey).To(Equal(expectedLRPKey))
				}

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			Context("when an auctioneer is present", func() {
				It("emits start auction requests", func() {
					Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))

					expectedStartRequest := auctioneer.NewLRPStartRequestFromModel(desiredLRP, 0, 1, 2, 3, 4)

					startAuctions := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
					Expect(startAuctions).To(HaveLen(1))
					Expect(startAuctions[0].ProcessGuid).To(Equal(desiredLRP.ProcessGuid))
					Expect(startAuctions[0].Indices).To(ConsistOf(expectedStartRequest.Indices))
				})
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.DesireLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})

			It("does not try to create actual LRPs", func() {
				Expect(fakeActualLRPDB.CreateUnclaimedActualLRPCallCount()).To(Equal(0))
			})
		})
	})

	Describe("UpdateDesiredLRP", func() {
		var (
			processGuid string
			update      *models.DesiredLRPUpdate

			requestBody interface{}
		)

		BeforeEach(func() {
			processGuid = "some-guid"
			someText := "some-text"
			update = &models.DesiredLRPUpdate{
				Annotation: &someText,
			}
			requestBody = &models.UpdateDesiredLRPRequest{
				ProcessGuid: processGuid,
				Update:      update,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.UpdateDesiredLRP(responseRecorder, request)
		})

		Context("when updating desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.UpdateDesiredLRPReturns(nil)
			})

			It("updates the desired lrp", func() {
				Expect(fakeDesiredLRPDB.UpdateDesiredLRPCallCount()).To(Equal(1))
				_, actualProcessGuid, actualUpdate := fakeDesiredLRPDB.UpdateDesiredLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(actualUpdate).To(Equal(update))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Error).To(BeNil())
			})

			Context("when the number of instances changes", func() {
				BeforeEach(func() {
					instances := int32(3)
					update.Instances = &instances

					desiredLRP := &models.DesiredLRP{
						ProcessGuid: "some-guid",
						Domain:      "some-domain",
						RootFs:      "some-stack",
						MemoryMb:    128,
						DiskMb:      512,
					}

					fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
					fakeServiceClient.CellByIdReturns(&models.CellPresence{RepAddress: "some-address"}, nil)
				})

				Context("when the number of instances decreased", func() {
					var actualLRPGroups []*models.ActualLRPGroup

					BeforeEach(func() {
						actualLRPGroups = []*models.ActualLRPGroup{}
						for i := 0; i < 5; i++ {
							actualLRPGroups = append(actualLRPGroups, &models.ActualLRPGroup{
								Instance: model_helpers.NewValidActualLRP("some-guid", int32(i)),
							})
						}
						fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(actualLRPGroups, nil)
					})

					It("stops extra actual lrps", func() {
						Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
						_, processGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("some-guid"))

						Expect(fakeActualLRPDB.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))
						_, processGuid = fakeActualLRPDB.ActualLRPGroupsByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("some-guid"))

						Expect(fakeServiceClient.CellByIdCallCount()).To(Equal(2))
						Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(2))
						Expect(fakeRepClientFactory.CreateClientArgsForCall(0)).To(Equal("some-address"))
						Expect(fakeRepClientFactory.CreateClientArgsForCall(1)).To(Equal("some-address"))

						Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(2))
						key, instanceKey := fakeRepClient.StopLRPInstanceArgsForCall(0)
						Expect(key).To(Equal(actualLRPGroups[3].Instance.ActualLRPKey))
						Expect(instanceKey).To(Equal(actualLRPGroups[3].Instance.ActualLRPInstanceKey))
						key, instanceKey = fakeRepClient.StopLRPInstanceArgsForCall(1)
						Expect(key).To(Equal(actualLRPGroups[4].Instance.ActualLRPKey))
						Expect(instanceKey).To(Equal(actualLRPGroups[4].Instance.ActualLRPInstanceKey))
					})

					Context("when fetching cell presence fails", func() {
						BeforeEach(func() {
							fakeServiceClient.CellByIdStub = func(lager.Logger, string) (*models.CellPresence, error) {
								if fakeRepClient.StopLRPInstanceCallCount() == 1 {
									return nil, errors.New("ohhhhh nooooo, mr billlll")
								} else {
									return &models.CellPresence{RepAddress: "some-address"}, nil
								}
							}
						})

						It("continues stopping the rest of the lrps and logs", func() {
							Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(1))
							Expect(logger).To(gbytes.Say("failed-fetching-cell-presence"))
						})
					})

					Context("when stopping the lrp fails", func() {
						BeforeEach(func() {
							fakeRepClient.StopLRPInstanceStub = func(models.ActualLRPKey, models.ActualLRPInstanceKey) error {
								if fakeRepClient.StopLRPInstanceCallCount() == 1 {
									return errors.New("ohhhhh nooooo, mr billlll")
								} else {
									return nil
								}
							}
						})

						It("continues stopping the rest of the lrps and logs", func() {
							Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(2))
							Expect(logger).To(gbytes.Say("failed-stopping-lrp-instance"))
						})
					})
				})

				Context("when the number of instances increases", func() {
					var runningActualLRPGroup *models.ActualLRPGroup

					BeforeEach(func() {
						runningActualLRPGroup = &models.ActualLRPGroup{
							Instance: model_helpers.NewValidActualLRP("some-guid", 0),
						}
						actualLRPGroups := []*models.ActualLRPGroup{
							runningActualLRPGroup,
						}
						fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(actualLRPGroups, nil)
					})

					It("creates missing actual lrps", func() {
						Expect(fakeDesiredLRPDB.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
						_, processGuid := fakeDesiredLRPDB.DesiredLRPByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("some-guid"))

						Expect(fakeActualLRPDB.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))
						_, processGuid = fakeActualLRPDB.ActualLRPGroupsByProcessGuidArgsForCall(0)
						Expect(processGuid).To(Equal("some-guid"))

						Expect(fakeActualLRPDB.CreateUnclaimedActualLRPCallCount()).To(Equal(2))
						_, key := fakeActualLRPDB.CreateUnclaimedActualLRPArgsForCall(0)
						Expect(key).To(BeEquivalentTo(&models.ActualLRPKey{
							ProcessGuid: "some-guid",
							Index:       1,
							Domain:      "some-domain",
						}))

						_, key = fakeActualLRPDB.CreateUnclaimedActualLRPArgsForCall(1)
						Expect(key).To(BeEquivalentTo(&models.ActualLRPKey{
							ProcessGuid: "some-guid",
							Index:       2,
							Domain:      "some-domain",
						}))

						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(1))
						startRequests := fakeAuctioneerClient.RequestLRPAuctionsArgsForCall(0)
						Expect(startRequests).To(BeEquivalentTo([]*auctioneer.LRPStartRequest{
							{ProcessGuid: "some-guid", Domain: "some-domain", Indices: []int{1, 2}, Resource: rep.Resource{MemoryMB: 128, DiskMB: 512, RootFs: "some-stack"}},
						}))
					})
				})

				Context("when fetching the desired lrp fails", func() {
					BeforeEach(func() {
						fakeDesiredLRPDB.DesiredLRPByProcessGuidReturns(nil, errors.New("you lose."))
					})

					It("does not update the actual lrps", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := models.DesiredLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeActualLRPDB.UnclaimActualLRPCallCount()).To(Equal(0))
						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})

				Context("when fetching the actual lrps groups fails", func() {
					BeforeEach(func() {
						fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(nil, errors.New("you lose."))
					})

					It("does not update the actual lrps", func() {
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := models.DesiredLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())
						Expect(response.Error).To(BeNil())

						Expect(fakeActualLRPDB.UnclaimActualLRPCallCount()).To(Equal(0))
						Expect(fakeAuctioneerClient.RequestLRPAuctionsCallCount()).To(Equal(0))
					})
				})
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.UpdateDesiredLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var (
			processGuid string

			requestBody interface{}
		)

		BeforeEach(func() {
			processGuid = "some-guid"
			requestBody = &models.RemoveDesiredLRPRequest{
				ProcessGuid: processGuid,
			}
		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.RemoveDesiredLRP(responseRecorder, request)
		})

		Context("when removing desired lrp in DB succeeds", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.RemoveDesiredLRPReturns(nil)
			})

			It("removes the desired lrp", func() {
				Expect(fakeDesiredLRPDB.RemoveDesiredLRPCallCount()).To(Equal(1))
				_, actualProcessGuid := fakeDesiredLRPDB.RemoveDesiredLRPArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})

			Context("when there are running instances on a present cell", func() {
				var runningActualLRPGroup, evacuatingAndRunningActualLRPGroup, evacuatingActualLRPGroup *models.ActualLRPGroup

				BeforeEach(func() {
					runningActualLRPGroup = &models.ActualLRPGroup{
						Instance: model_helpers.NewValidActualLRP("some-guid", 0),
					}
					evacuatingAndRunningActualLRPGroup = &models.ActualLRPGroup{
						Instance:   model_helpers.NewValidActualLRP("some-guid", 1),
						Evacuating: model_helpers.NewValidActualLRP("some-guid", 1),
					}
					evacuatingActualLRPGroup = &models.ActualLRPGroup{
						Evacuating: model_helpers.NewValidActualLRP("some-guid", 2),
					}

					actualLRPGroups := []*models.ActualLRPGroup{
						runningActualLRPGroup,
						evacuatingAndRunningActualLRPGroup,
						evacuatingActualLRPGroup,
					}

					fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(actualLRPGroups, nil)
				})

				It("stops all of the corresponding actual lrps", func() {
					Expect(fakeActualLRPDB.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))

					_, processGuid := fakeActualLRPDB.ActualLRPGroupsByProcessGuidArgsForCall(0)
					Expect(processGuid).To(Equal("some-guid"))

					Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(2))
					Expect(fakeRepClientFactory.CreateClientArgsForCall(0)).To(Equal(runningActualLRPGroup.Instance.CellId))
					Expect(fakeRepClientFactory.CreateClientArgsForCall(1)).To(Equal(evacuatingAndRunningActualLRPGroup.Instance.CellId))

					Expect(fakeRepClient.StopLRPInstanceCallCount()).To(Equal(2))
					key, instanceKey := fakeRepClient.StopLRPInstanceArgsForCall(0)
					Expect(key).To(Equal(runningActualLRPGroup.Instance.ActualLRPKey))
					Expect(instanceKey).To(Equal(runningActualLRPGroup.Instance.ActualLRPInstanceKey))
					key, instanceKey = fakeRepClient.StopLRPInstanceArgsForCall(1)
					Expect(key).To(Equal(evacuatingAndRunningActualLRPGroup.Instance.ActualLRPKey))
					Expect(instanceKey).To(Equal(evacuatingAndRunningActualLRPGroup.Instance.ActualLRPInstanceKey))
				})

				Context("when fetching the actual lrps fails", func() {
					BeforeEach(func() {
						fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(nil, errors.New("new error dawg"))
					})

					It("logs the error but still succeeds", func() {
						Expect(fakeRepClientFactory.CreateClientCallCount()).To(Equal(0))
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
						response := models.DesiredLRPLifecycleResponse{}
						err := response.Unmarshal(responseRecorder.Body.Bytes())
						Expect(err).NotTo(HaveOccurred())

						Expect(response.Error).To(BeNil())
						Expect(logger).To(gbytes.Say("failed-fetching-actual-lrps"))
					})
				})
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeDesiredLRPDB.RemoveDesiredLRPReturns(models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := models.DesiredLRPLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
			})
		})
	})
})
