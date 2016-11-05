package bbs_test

import (
	"os"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/locket"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/ifrit/grouper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceClient", func() {
	var serviceClient bbs.ServiceClient

	BeforeEach(func() {
		serviceClient = bbs.NewServiceClient(consulClient, clock.NewClock())
	})

	Describe("CellById", func() {
		const cellID = "cell-id"

		Context("when the cell exists", func() {
			It("returns the correct CellPresence", func() {
				cellPresence := newCellPresence(cellID)
				consulHelper.RegisterCell(cellPresence)
				presence, err := serviceClient.CellById(logger, cellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(presence).To(BeEquivalentTo(cellPresence))
			})
		})

		Context("when the cell does not exist", func() {
			It("returns ErrStoreResourceNotFound", func() {
				_, err := serviceClient.CellById(logger, cellID)
				Expect(err).To(HaveOccurred())
				modelErr := models.ConvertError(err)
				Expect(modelErr.Type).To(Equal(models.Error_ResourceNotFound))
			})
		})
	})

	Describe("Cells", func() {
		const cell1 = "cell-id-1"
		const cell2 = "cell-id-2"

		Context("when there is a single cell", func() {
			var maintainers ifrit.Process

			BeforeEach(func() {
				Expect(serviceClient.Cells(logger)).To(HaveLen(0))
				maintainers = ifrit.Invoke(grouper.NewParallel(os.Interrupt, grouper.Members{
					{cell1, serviceClient.NewCellPresenceRunner(logger, newCellPresence(cell1), locket.RetryInterval, locket.DefaultSessionTTL)},
					{cell2, serviceClient.NewCellPresenceRunner(logger, newCellPresence(cell2), locket.RetryInterval, locket.DefaultSessionTTL)},
				}))
			})

			AfterEach(func() {
				ginkgomon.Interrupt(maintainers)
			})

			It("returns only one cell", func() {
				Eventually(func() (models.CellSet, error) { return serviceClient.Cells(logger) }).Should(HaveLen(2))
				Expect(serviceClient.Cells(logger)).To(HaveKey(cell1))
				Expect(serviceClient.Cells(logger)).To(HaveKey(cell2))
			})
		})
	})
})

func newCellPresence(cellID string) *models.CellPresence {
	presence := models.NewCellPresence(
		cellID,
		"cell.example.com",
		"http://cell.example.com",
		"the-zone",
		models.NewCellCapacity(128, 1024, 6),
		[]string{},
		nil,
		nil,
		nil,
	)
	return &presence
}
