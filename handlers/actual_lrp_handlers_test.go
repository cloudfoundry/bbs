package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/db/fakes"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("ActualLRP Handlers", func() {
	var (
		logger           lager.Logger
		fakeActualLRPDB  *fakes.FakeActualLRPDB
		responseRecorder *httptest.ResponseRecorder
		handler          *handlers.ActualLRPHandler

		actualLRP1     models.ActualLRP
		actualLRP2     models.ActualLRP
		evacuatingLRP2 models.ActualLRP
	)

	BeforeEach(func() {
		fakeActualLRPDB = new(fakes.FakeActualLRPDB)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewActualLRPHandler(fakeActualLRPDB, logger)
	})

	Describe("ActualLRPGroups", func() {
		var request *http.Request

		BeforeEach(func() {
			request = newTestRequest("")

			actualLRP1 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-0",
					1,
					"domain-0",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-0",
					"cell-id-0",
				),
				State: proto.String(models.ActualLRPStateRunning),
				Since: proto.Int64(1138),
			}

			actualLRP2 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-1",
					2,
					"domain-1",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-1",
					"cell-id-1",
				),
				State: proto.String(models.ActualLRPStateClaimed),
				Since: proto.Int64(4444),
			}

			evacuatingLRP2 = actualLRP2
			evacuatingLRP2.State = proto.String(models.ActualLRPStateRunning)
			evacuatingLRP2.Since = proto.Int64(3417)
		})

		JustBeforeEach(func() {
			handler.ActualLRPGroups(responseRecorder, request)
		})

		Context("when reading actual lrps from DB succeeds", func() {
			var actualLRPGroups *models.ActualLRPGroups

			BeforeEach(func() {
				actualLRPGroups = &models.ActualLRPGroups{
					[]*models.ActualLRPGroup{
						{Instance: &actualLRP1},
						{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
					},
				}
				fakeActualLRPDB.ActualLRPGroupsReturns(actualLRPGroups, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp groups", func() {
				response := &models.ActualLRPGroups{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(actualLRPGroups))
			})

			Context("and no filter is provided", func() {
				It("call the DB with no filters to retrieve the actual lrp groups", func() {
					Expect(fakeActualLRPDB.ActualLRPGroupsCallCount()).To(Equal(1))
					filter, _ := fakeActualLRPDB.ActualLRPGroupsArgsForCall(0)
					Expect(filter).To(Equal(models.ActualLRPFilter{}))
				})
			})

			Context("and filtering by domain", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?domain=domain-1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("call the DB with the domain filter to retrieve the actual lrp groups", func() {
					Expect(fakeActualLRPDB.ActualLRPGroupsCallCount()).To(Equal(1))
					filter, _ := fakeActualLRPDB.ActualLRPGroupsArgsForCall(0)
					Expect(filter.Domain).To(Equal("domain-1"))
				})
			})

			Context("and filtering by cellId", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?cell_id=cellid-1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("call the DB with the cell id filter to retrieve the actual lrp groups", func() {
					Expect(fakeActualLRPDB.ActualLRPGroupsCallCount()).To(Equal(1))
					filter, _ := fakeActualLRPDB.ActualLRPGroupsArgsForCall(0)
					Expect(filter.CellID).To(Equal("cellid-1"))
				})
			})

			Context("and filtering by cellId and domain", func() {
				BeforeEach(func() {
					var err error
					request, err = http.NewRequest("", "http://example.com?domain=potato&cell_id=cellid-1", nil)
					Expect(err).NotTo(HaveOccurred())
				})

				It("call the DB with the both filters to retrieve the actual lrp groups", func() {
					Expect(fakeActualLRPDB.ActualLRPGroupsCallCount()).To(Equal(1))
					filter, _ := fakeActualLRPDB.ActualLRPGroupsArgsForCall(0)
					Expect(filter.CellID).To(Equal("cellid-1"))
					Expect(filter.Domain).To(Equal("potato"))
				})
			})
		})

		Context("when the DB returns no actual lrp groups", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupsReturns(&models.ActualLRPGroups{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := &models.ActualLRPGroups{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(&models.ActualLRPGroups{}))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupsReturns(&models.ActualLRPGroups{}, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})

	Describe("ActualLRPGroupsByProcessGuid", func() {
		var request *http.Request
		var processGuid = "process-guid"

		BeforeEach(func() {
			request = newTestRequest("")
			request.URL.RawQuery = url.Values{":process_guid": []string{processGuid}}.Encode()

			actualLRP1 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-0",
					"cell-id-0",
				),
				State: proto.String(models.ActualLRPStateRunning),
				Since: proto.Int64(1138),
			}

			actualLRP2 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-1",
					2,
					"domain-1",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-1",
					"cell-id-1",
				),
				State: proto.String(models.ActualLRPStateClaimed),
				Since: proto.Int64(4444),
			}

			evacuatingLRP2 = actualLRP2
			evacuatingLRP2.State = proto.String(models.ActualLRPStateRunning)
			evacuatingLRP2.Since = proto.Int64(3417)
		})

		JustBeforeEach(func() {
			handler.ActualLRPGroupsByProcessGuid(responseRecorder, request)
		})

		Context("when reading actual lrps from DB succeeds", func() {
			var actualLRPGroups *models.ActualLRPGroups

			BeforeEach(func() {
				actualLRPGroups = &models.ActualLRPGroups{
					[]*models.ActualLRPGroup{
						{Instance: &actualLRP1},
						{Instance: &actualLRP2, Evacuating: &evacuatingLRP2},
					},
				}
				fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(actualLRPGroups, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("fetches actual lrp groups by process guid", func() {
				Expect(fakeActualLRPDB.ActualLRPGroupsByProcessGuidCallCount()).To(Equal(1))
				actualProcessGuid, _ := fakeActualLRPDB.ActualLRPGroupsByProcessGuidArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns a list of actual lrp groups", func() {
				response := &models.ActualLRPGroups{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(actualLRPGroups))
			})
		})

		Context("when the DB returns no actual lrp groups", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(&models.ActualLRPGroups{}, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("returns an empty list", func() {
				response := &models.ActualLRPGroups{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(&models.ActualLRPGroups{}))
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupsByProcessGuidReturns(&models.ActualLRPGroups{}, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})

	Describe("ActualLRPGroupByProcessGuidAndIndex", func() {
		var request *http.Request
		var processGuid = "process-guid"
		var index = 1

		BeforeEach(func() {
			request = newTestRequest("")
			request.URL.RawQuery = url.Values{
				":process_guid": []string{processGuid},
				":index":        []string{strconv.Itoa(index)},
			}.Encode()

			actualLRP1 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					processGuid,
					1,
					"domain-0",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-0",
					"cell-id-0",
				),
				State: proto.String(models.ActualLRPStateRunning),
				Since: proto.Int64(1138),
			}

			actualLRP2 = models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey(
					"process-guid-1",
					2,
					"domain-1",
				),
				ActualLRPInstanceKey: models.NewActualLRPInstanceKey(
					"instance-guid-1",
					"cell-id-1",
				),
				State: proto.String(models.ActualLRPStateClaimed),
				Since: proto.Int64(4444),
			}

			evacuatingLRP2 = actualLRP2
			evacuatingLRP2.State = proto.String(models.ActualLRPStateRunning)
			evacuatingLRP2.Since = proto.Int64(3417)
		})

		JustBeforeEach(func() {
			handler.ActualLRPGroupByProcessGuidAndIndex(responseRecorder, request)
		})

		Context("when reading actual lrps from DB succeeds", func() {
			var actualLRPGroup *models.ActualLRPGroup

			BeforeEach(func() {
				actualLRPGroup = &models.ActualLRPGroup{Instance: &actualLRP1}
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			It("responds with 200 Status OK", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})

			It("fetches actual lrp group by process guid and index", func() {
				Expect(fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))
				actualProcessGuid, idx, _ := fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
				Expect(actualProcessGuid).To(Equal(processGuid))
				Expect(idx).To(BeEquivalentTo(index))
			})

			It("returns an actual lrp group", func() {
				response := &models.ActualLRPGroup{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(actualLRPGroup))
			})

			Context("when there is also an evacuating LRP", func() {
				BeforeEach(func() {
					actualLRPGroup = &models.ActualLRPGroup{Instance: &actualLRP2, Evacuating: &evacuatingLRP2}
					fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
				})

				It("returns both LRPs in the group", func() {
					response := &models.ActualLRPGroup{}
					err := response.Unmarshal(responseRecorder.Body.Bytes())
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(Equal(actualLRPGroup))
				})
			})
		})

		Context("when we cannot find the resource", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, models.ErrResourceNotFound)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrResourceNotFound)).To(BeTrue())
			})
		})

		Context("when the DB errors out", func() {
			BeforeEach(func() {
				fakeActualLRPDB.ActualLRPGroupByProcessGuidAndIndexReturns(nil, models.ErrUnknownError)
			})

			It("responds with an error", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("provides relevant error information", func() {
				var bbsError models.Error
				err := bbsError.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(bbsError.Equal(models.ErrUnknownError)).To(BeTrue())
			})
		})
	})
})
