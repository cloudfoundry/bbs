package consul_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cell DB", func() {
	Describe("CellById", func() {
		var (
			cellPresence models.CellPresence
			cellID       string
		)

		BeforeEach(func() {
			cellID = "cell-id"
			cellPresence = models.NewCellPresence(
				cellID,
				"cell.example.com",
				"the-zone",
				models.NewCellCapacity(128, 1024, 6),
				[]string{},
				[]string{},
			)
		})

		Context("when the cell exists", func() {
			BeforeEach(func() {
				consulHelper.RegisterCell(cellPresence)
			})

			It("returns the correct CellPresence", func() {
				presence, err := consulDB.CellById(logger, cellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(*presence).To(Equal(cellPresence))
			})
		})

		Context("when the cell does not exist", func() {
			It("returns ErrStoreResourceNotFound", func() {
				_, err := consulDB.CellById(logger, cellID)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
