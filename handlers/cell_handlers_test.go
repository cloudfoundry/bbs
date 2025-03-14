package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/serviceclient/serviceclientfakes"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("Cell Handlers", func() {
	var (
		logger            *lagertest.TestLogger
		responseRecorder  *httptest.ResponseRecorder
		handler           *handlers.CellHandler
		fakeServiceClient *serviceclientfakes.FakeServiceClient
		exitCh            chan struct{}
		cells             []*models.CellPresence
		cellSet           models.CellSet

		requestIdHeader   string
		b3RequestIdHeader string
	)

	BeforeEach(func() {
		fakeServiceClient = new(serviceclientfakes.FakeServiceClient)
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()
		exitCh = make(chan struct{}, 1)
		requestIdHeader = "0bc29108-c522-4360-93dd-30ca38cce13d"
		b3RequestIdHeader = fmt.Sprintf(`"trace-id":"%s"`, strings.Replace(requestIdHeader, "-", "", -1))
		handler = handlers.NewCellHandler(fakeServiceClient, exitCh)
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
				RootfsProviders: []*models.Provider{
					&models.Provider{Name: "preloaded", Properties: []string{"provider-1", "provider-2"}},
					&models.Provider{Name: "provider-3", Properties: nil},
				},
				PlacementTags: []string{"test1", "test2"},
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
				RootfsProviders: []*models.Provider{
					&models.Provider{Name: "preloaded", Properties: []string{"provider-1"}},
				},
				PlacementTags: []string{"test3", "test4"},
			},
		}
		cellSet = models.NewCellSet()
		cellSet.Add(cells[0])
		cellSet.Add(cells[1])
	})

	Describe("Cells", func() {
		JustBeforeEach(func() {
			request := newTestRequest("")
			request.Header.Set(lager.RequestIdHeader, requestIdHeader)
			handler.Cells(logger, responseRecorder, request)
		})

		Context("when reading cells succeeds", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(cellSet, nil)
			})

			It("returns a list of cells", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.CellsResponse
				var protoResponse models.ProtoCellsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
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

				var response models.CellsResponse
				var protoResponse models.ProtoCellsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
				Expect(response.Cells).To(BeNil())
			})
		})

		Context("when the DB returns an unrecoverable error", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(nil, models.NewUnrecoverableError(nil))
			})

			It("logs and writes to the exit channel", func() {
				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
				Eventually(logger).Should(gbytes.Say(b3RequestIdHeader))
				Eventually(exitCh).Should(Receive())
			})
		})

		Context("when the serviceClient errors out", func() {
			BeforeEach(func() {
				fakeServiceClient.CellsReturns(nil, models.ErrUnknownError)
			})

			It("provides relevant error information", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))

				var response models.CellsResponse
				var protoResponse models.ProtoCellsResponse
				err := proto.Unmarshal(responseRecorder.Body.Bytes(), &protoResponse)
				response = *protoResponse.FromProto()
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(Equal(models.ErrUnknownError))
				Expect(response.Cells).To(BeNil())
			})
		})
	})
})
