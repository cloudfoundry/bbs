package consul_test

import (
	"time"

	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/locket/presence"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
)

var _ = Describe("CellsLoader", func() {
	Describe("Cells", func() {

		const ttl = 10 * time.Second
		const retryInterval = time.Second
		var (
			clock *fakeclock.FakeClock

			locketClient       locket.Client
			presence1          ifrit.Process
			presence2          ifrit.Process
			firstCellPresence  presence.CellPresence
			secondCellPresence presence.CellPresence
			logger             *lagertest.TestLogger
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test")
			clock = fakeclock.NewFakeClock(time.Now())
			locketClient = locket.NewClient(consulSession, clock, logger)

			firstCellPresence = presence.NewCellPresence("first-rep", "1.2.3.4", "the-zone", presence.NewCellCapacity(128, 1024, 3), []string{}, []string{})
			secondCellPresence = presence.NewCellPresence("second-rep", "4.5.6.7", "the-zone", presence.NewCellCapacity(128, 1024, 3), []string{}, []string{})

			presence1 = nil
			presence2 = nil

		})

		AfterEach(func() {
			ginkgomon.Interrupt(presence1)
			ginkgomon.Interrupt(presence2)
		})

		Context("when there is a single cell", func() {
			var cellsLoader db.CellsLoader
			var cells models.CellSet
			var err error

			BeforeEach(func() {
				cellsLoader = consulDB.NewCellsLoader(logger)
				presence1 = ifrit.Invoke(locketClient.NewCellPresence(firstCellPresence, retryInterval))

				Eventually(func() ([]presence.CellPresence, error) {
					return locketClient.Cells()
				}).Should(HaveLen(1))

				cells, err = cellsLoader.Cells()
			})

			It("returns only one cell", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(cells).To(HaveLen(1))
				Expect(cells).To(HaveKey("first-rep"))
			})

			Context("when one more cell is added", func() {
				BeforeEach(func() {
					presence2 = ifrit.Invoke(locketClient.NewCellPresence(secondCellPresence, retryInterval))

					Eventually(func() ([]presence.CellPresence, error) {
						return locketClient.Cells()
					}).Should(HaveLen(2))
				})

				It("returns only one cell", func() {
					cells, err := cellsLoader.Cells()
					Expect(err).NotTo(HaveOccurred())
					Expect(cells).To(HaveLen(1))
				})

				Context("when a new loader is created", func() {
					It("returns two cells", func() {
						newCellsLoader := consulDB.NewCellsLoader(logger)
						cells, err := newCellsLoader.Cells()
						Expect(err).NotTo(HaveOccurred())
						Expect(cells).To(HaveLen(2))
						Expect(cells).To(HaveKey("first-rep"))
						Expect(cells).To(HaveKey("second-rep"))
					})
				})
			})
		})
	})
})
