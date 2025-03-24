package serviceclient_test

import (
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/lager/v3/lagertest"
	locketmodels "code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/locket/models/modelsfakes"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceClient", func() {
	var (
		locketClient                 *modelsfakes.FakeLocketClient
		serviceClient                serviceclient.ServiceClient
		logger                       *lagertest.TestLogger
		cellPresence1, cellPresence2 *models.CellPresence
		cellPresence3                *models.CellPresence
	)

	resourceFromPresence := func(presence *models.CellPresence) *locketmodels.Resource {
		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())

		data, err := json.Marshal(presence)
		Expect(err).NotTo(HaveOccurred())

		return &locketmodels.Resource{
			Key:   presence.CellId,
			Owner: guid.String(),
			Value: string(data),
		}
	}

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("service-client")

		locketClient = &modelsfakes.FakeLocketClient{}

		cellPresence1 = &models.CellPresence{
			CellId:     "cell-1",
			RepAddress: "cell-1-address",
		}
		cellPresence2 = &models.CellPresence{
			CellId:     "cell-2",
			RepAddress: "cell-2-address",
		}
		cellPresence3 = &models.CellPresence{
			CellId:     "cell-3",
			RepAddress: "cell-3-address",
		}

		serviceClient = serviceclient.NewServiceClient(locketClient)
	})

	Context("Cells", func() {
		BeforeEach(func() {
			locketClient.FetchAllReturns(&locketmodels.FetchAllResponse{
				Resources: []*locketmodels.Resource{
					resourceFromPresence(cellPresence1),
					resourceFromPresence(cellPresence2),
					resourceFromPresence(cellPresence3),
				},
			}, nil)
		})

		It("fetches the cells from the locket client", func() {
			set, err := serviceClient.Cells(logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(set).To(HaveLen(3))

			Expect(set["cell-1"]).To(Equal(cellPresence1))
			Expect(set["cell-2"]).To(Equal(cellPresence2))
			Expect(set["cell-3"]).To(Equal(cellPresence3))

			Expect(locketClient.FetchAllCallCount()).To(Equal(1))
			_, request, _ := locketClient.FetchAllArgsForCall(0)
			Expect(request).To(Equal(&locketmodels.FetchAllRequest{Type: locketmodels.PresenceType, TypeCode: locketmodels.PRESENCE}))
		})

		Context("when fetching the cells from the locket client fails", func() {
			BeforeEach(func() {
				locketClient.FetchAllReturns(nil, errors.New("boom"))
			})

			It("returns an error", func() {
				_, err := serviceClient.Cells(logger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the cell presence has invalid data", func() {
			BeforeEach(func() {
				locketClient.FetchAllReturns(&locketmodels.FetchAllResponse{
					Resources: []*locketmodels.Resource{
						resourceFromPresence(cellPresence1),
						resourceFromPresence(cellPresence2),
						resourceFromPresence(cellPresence3),
						&locketmodels.Resource{
							Key:   "cell-4",
							Value: "{{",
						},
					},
				}, nil)
			})

			It("ignores the cell", func() {
				set, err := serviceClient.Cells(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(set).To(HaveLen(3))

				Expect(set["cell-1"]).To(Equal(cellPresence1))
				Expect(set["cell-2"]).To(Equal(cellPresence2))
				Expect(set["cell-3"]).To(Equal(cellPresence3))
			})
		})
	})

	Context("CellById", func() {
		BeforeEach(func() {
			locketClient.FetchReturns(&locketmodels.FetchResponse{
				Resource: resourceFromPresence(cellPresence2),
			}, nil)
		})

		It("fetches the cell presence", func() {
			presence, err := serviceClient.CellById(logger, "cell-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(presence).To(Equal(cellPresence2))

			Expect(locketClient.FetchCallCount()).To(Equal(1))
			_, request, _ := locketClient.FetchArgsForCall(0)
			Expect(request).To(Equal(&locketmodels.FetchRequest{Key: "cell-1"}))
		})

		Context("when the cell presence does not exist", func() {
			BeforeEach(func() {
				locketClient.FetchReturns(nil, locketmodels.ErrResourceNotFound)
			})

			It("returns an error", func() {
				_, err := serviceClient.CellById(logger, "cell-1")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when the locket presence has invalid data", func() {
			BeforeEach(func() {
				locketClient.FetchReturns(&locketmodels.FetchResponse{
					Resource: &locketmodels.Resource{
						Key:   "foobar",
						Value: "{{",
					},
				}, nil)
			})

			It("returns an error", func() {
				_, err := serviceClient.CellById(logger, "cell-1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the cell presence client is nil", func() {
			BeforeEach(func() {
				serviceClient = serviceclient.NewServiceClient(locketClient)
			})

			It("fetches the cell presence from ", func() {
				presence, err := serviceClient.CellById(logger, "cell-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(presence).To(Equal(cellPresence2))

				Expect(locketClient.FetchCallCount()).To(Equal(1))
				_, request, _ := locketClient.FetchArgsForCall(0)
				Expect(request).To(Equal(&locketmodels.FetchRequest{Key: "cell-1"}))
			})

			Context("and cell is not found in locket", func() {
				BeforeEach(func() {
					locketClient.FetchReturns(nil, locketmodels.ErrResourceNotFound)
				})

				It("returns a ResourceNotFound error", func() {
					_, err := serviceClient.CellById(logger, "cell-1")
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(models.ErrResourceNotFound))
				})
			})
		})
	})

	Context("CellEvents", func() {
		BeforeEach(func() {
			serviceClient = serviceclient.NewServiceClient(locketClient)
		})

		It("returns nil", func() {
			events := serviceClient.CellEvents(logger)
			Expect(events).To(BeNil())
		})
	})
})
