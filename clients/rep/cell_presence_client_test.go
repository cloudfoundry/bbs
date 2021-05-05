package rep_test

import (
	"os"

	"code.cloudfoundry.org/bbs/clients/rep"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/tedsuo/ifrit/grouper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CellPresenceClient", func() {
	var (
		cellPresenceClient rep.CellPresenceClient
		logger             *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("cell-presence-client")
		cellPresenceClient = rep.NewCellPresenceClient(consulClient, clock.NewClock())
	})

	Describe("CellById", func() {
		var (
			cellID       string
			process      ifrit.Process
			cellPresence *models.CellPresence
		)

		BeforeEach(func() {
			cellID = "cell-id"
		})

		Context("when the cell exists", func() {
			BeforeEach(func() {
				cellPresence = newCellPresence(cellID)
				process = ifrit.Invoke(cellPresenceClient.NewCellPresenceRunner(logger, newCellPresence(cellID), locket.RetryInterval, locket.DefaultSessionTTL))
			})

			AfterEach(func() {
				ginkgomon.Interrupt(process)
			})

			It("returns the correct CellPresence", func() {
				presence, err := cellPresenceClient.CellById(logger, cellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(presence).To(BeEquivalentTo(cellPresence))
			})
		})

		Context("when the cell does not exist", func() {
			It("returns ErrStoreResourceNotFound", func() {
				_, err := cellPresenceClient.CellById(logger, cellID)
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
				Expect(cellPresenceClient.Cells(logger)).To(HaveLen(0))
				maintainers = ifrit.Invoke(grouper.NewParallel(os.Interrupt, grouper.Members{
					{cell1, cellPresenceClient.NewCellPresenceRunner(logger, newCellPresence(cell1), locket.RetryInterval, locket.DefaultSessionTTL)},
					{cell2, cellPresenceClient.NewCellPresenceRunner(logger, newCellPresence(cell2), locket.RetryInterval, locket.DefaultSessionTTL)},
				}))
			})

			AfterEach(func() {
				ginkgomon.Interrupt(maintainers)
			})

			It("returns only one cell", func() {
				Eventually(func() (models.CellSet, error) { return cellPresenceClient.Cells(logger) }).Should(HaveLen(2))
				Expect(cellPresenceClient.Cells(logger)).To(HaveKey(cell1))
				Expect(cellPresenceClient.Cells(logger)).To(HaveKey(cell2))
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
