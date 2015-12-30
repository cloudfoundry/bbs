package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/bbs/fake_bbs"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("Cell Handlers", func() {
	var (
		logger            lager.Logger
		responseRecorder  *httptest.ResponseRecorder
		handler           *handlers.CellHandler
		fakeServiceClient *fake_bbs.FakeServiceClient
	)

	BeforeEach(func() {
		fakeServiceClient = new(fake_bbs.FakeServiceClient)
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		responseRecorder = httptest.NewRecorder()
		handler = handlers.NewCellHandler(logger, fakeServiceClient)
	})

	Describe("Cells", func() {
		var cells []*models.CellPresence
		var cellSet models.CellSet
		BeforeEach(func() {
			cells = []*models.CellPresence{
				{
					CellId:     "cell-1",
					RepAddress: "1.1.1.1",
					Zone:       "z1",
					Capacity: &models.CellCapacity{
						MemoryMb:   1000,
						DiskMb:     1000,
						Containers: 50,
					},
					RootfsProviders: models.RootFSProviders{
						"provider1": &models.Providers{ProvidersList: []string{"test1", "test2"}},
						"provider2": &models.Providers{ProvidersList: []string{"test3", "test4"}},
					},
				},
				{
					CellId:     "cell-2",
					RepAddress: "2.2.2.2",
					Zone:       "z2",
					Capacity: &models.CellCapacity{
						MemoryMb:   2000,
						DiskMb:     2000,
						Containers: 20,
					},
					RootfsProviders: models.RootFSProviders{
						"provider1": &models.Providers{ProvidersList: []string{"test1", "test2"}},
						"provider2": &models.Providers{ProvidersList: []string{"test3", "test4"}},
					},
				},
			}
			cellSet = models.NewCellSet()
			cellSet.Add(cells[0])
			cellSet.Add(cells[1])
		})

		JustBeforeEach(func() {
			handler.Cells(responseRecorder, newTestRequest(""))
		})

		Context("when reading cells succeeds", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(cellSet, nil)
			})

			It("returns a list of cells", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := &models.CellsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Cells).To(ConsistOf(cells))
			})
		})

		Context("when the serviceClient returns no cells", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(nil, nil)
			})

			It("returns an empty list", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := &models.CellsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Cells).To(BeNil())
			})
		})

		Context("when the serviceClient errors out", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				response := &models.CellsResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
				Expect(response.Cells).To(BeNil())
			})
		})
	})
})
