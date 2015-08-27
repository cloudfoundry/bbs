package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Convergence API", func() {
	Describe("ConvergeLRPs", func() {
		var processGuid string

		BeforeEach(func() {
			cellPresence := models.NewCellPresence(
				"some-cell",
				"cell.example.com",
				"the-zone",
				models.NewCellCapacity(128, 1024, 6),
				[]string{},
				[]string{},
			)
			consulHelper.RegisterCell(cellPresence)
			processGuid = "some-process-guid"
			etcdHelper.CreateValidDesiredLRP(processGuid)
		})

		It("converges the lrps", func() {
			err := client.ConvergeLRPs()
			Expect(err).NotTo(HaveOccurred())

			groups, err := client.ActualLRPGroupsByProcessGuid(processGuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(groups).NotTo(HaveLen(0))
		})
	})
})
