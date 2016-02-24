package sqldb_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRPDB", func() {
	Describe("DesireLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP("the-guid")
		})

		It("saves the lrp in the database", func() {
			err := sqlDB.DesireLRP(logger, expectedDesiredLRP)
			Expect(err).NotTo(HaveOccurred())

			var desiredLRP models.DesiredLRP
			desiredLRP.ModificationTag = &models.ModificationTag{}
			var routeData, runInformationData []byte

			row := db.QueryRow("SELECT * FROM desired_lrps WHERE process_guid = ?", expectedDesiredLRP.ProcessGuid)
			err = row.Scan(
				&desiredLRP.ProcessGuid,
				&desiredLRP.Domain,
				&desiredLRP.LogGuid,
				&desiredLRP.Annotation,
				&desiredLRP.Instances,
				&desiredLRP.MemoryMb,
				&desiredLRP.DiskMb,
				&desiredLRP.RootFs,
				&routeData,
				&desiredLRP.ModificationTag.Epoch,
				&desiredLRP.ModificationTag.Index,
				&runInformationData,
			)
			Expect(err).NotTo(HaveOccurred())

			var runInformation models.DesiredLRPRunInfo
			err = serializer.Unmarshal(logger, runInformationData, &runInformation)
			Expect(err).NotTo(HaveOccurred())
			desiredLRP.AddRunInfo(runInformation)

			var routes models.Routes
			err = json.Unmarshal(routeData, &routes)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP.Routes = &routes

			Expect(desiredLRP.Equal(expectedDesiredLRP)).To(BeTrue())
		})

		Context("when serializing the run information fails", func() {
			BeforeEach(func() {
				expectedDesiredLRP.CpuWeight = 1000
			})

			It("returns a bad request error", func() {
				err := sqlDB.DesireLRP(logger, expectedDesiredLRP)
				modelErr := models.ConvertError(err)
				Expect(modelErr).NotTo(BeNil())
				Expect(modelErr.Type).To(Equal(models.Error_InvalidRecord))
			})
		})

		Context("when the process_guid is already taken", func() {
			BeforeEach(func() {
				err := sqlDB.DesireLRP(logger, expectedDesiredLRP)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a resource exists error", func() {
				err := sqlDB.DesireLRP(logger, expectedDesiredLRP)
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})
	})
})
