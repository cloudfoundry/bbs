package main_test

import (
	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Convergence API", func() {
	Describe("ConvergeLRPs", func() {
		var processGuid string

		BeforeEach(func() {
			bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
			bbsProcess = ginkgomon.Invoke(bbsRunner)

			cellPresence := models.NewCellPresence(
				"some-cell",
				"cell.example.com",
				"the-zone",
				models.NewCellCapacity(128, 1024, 6),
				[]string{},
				[]string{},
			)
			consulHelper.RegisterCell(&cellPresence)
			processGuid = "some-process-guid"
			err := client.DesireLRP(model_helpers.NewValidDesiredLRP(processGuid))
			Expect(err).NotTo(HaveOccurred())
			err = client.RemoveActualLRP(processGuid, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ginkgomon.Kill(bbsProcess)
		})

		It("converges the lrps", func() {
			err := client.ConvergeLRPs()
			Expect(err).NotTo(HaveOccurred())

			groups, err := client.ActualLRPGroupsByProcessGuid(processGuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(groups).To(HaveLen(1))
		})
	})
})
