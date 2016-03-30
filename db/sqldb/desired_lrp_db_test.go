package sqldb_test

import (
	"encoding/json"
	"fmt"

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

			Expect(desiredLRP).To(Equal(*expectedDesiredLRP))
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

	Describe("DesiredLRPByProcessGuid", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(logger, expectedDesiredLRP)).To(Succeed())
		})

		It("returns the desired lrp", func() {
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		Context("when the desired lrp does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, "Sup dawg")
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(desiredLRP).To(BeNil())
			})
		})

		Context("when the run info is invalid", func() {
			BeforeEach(func() {
				result, err := db.Exec("UPDATE desired_lrps SET run_info = ? WHERE process_guid = ?", "{{", expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("returns an invalid record error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(desiredLRP).To(BeNil())
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				result, err := db.Exec("UPDATE desired_lrps SET routes = ? WHERE process_guid = ?", "{{", expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("returns an invalid record error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(desiredLRP).To(BeNil())
			})
		})
	})

	Describe("DesiredLRPs", func() {
		var expectedDesiredLRPs []*models.DesiredLRP

		BeforeEach(func() {
			expectedDesiredLRPs = []*models.DesiredLRP{}
			expectedDesiredLRPs = append(expectedDesiredLRPs, model_helpers.NewValidDesiredLRP("d-1"))
			expectedDesiredLRPs = append(expectedDesiredLRPs, model_helpers.NewValidDesiredLRP("d-2"))
			for i, expectedDesiredLRP := range expectedDesiredLRPs {
				expectedDesiredLRP.Domain = fmt.Sprintf("domain-%d", i+1)
				Expect(sqlDB.DesireLRP(logger, expectedDesiredLRP)).To(Succeed())
			}
		})

		It("returns all desired lrps", func() {
			desiredLRPs, err := sqlDB.DesiredLRPs(logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPs).To(HaveLen(2))
			Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
		})

		Context("when filtering by domain", func() {
			It("returns a ResourceNotFound error", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(logger, models.DesiredLRPFilter{Domain: "domain-1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(HaveLen(1))
				Expect(desiredLRPs[0]).To(BeEquivalentTo(expectedDesiredLRPs[0]))
			})
		})

		Context("when the run info is invalid", func() {
			BeforeEach(func() {
				result, err := db.Exec("UPDATE desired_lrps SET run_info = ? WHERE process_guid = ?", "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(HaveLen(1))
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				result, err := db.Exec("UPDATE desired_lrps SET routes = ? WHERE process_guid = ?", "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(HaveLen(1))
			})
		})
	})

	Describe("DesiredLRPSchedulingInfos", func() {
		var expectedDesiredLRPSchedulingInfos []*models.DesiredLRPSchedulingInfo
		var expectedDesiredLRPs []*models.DesiredLRP

		BeforeEach(func() {
			expectedDesiredLRPs = []*models.DesiredLRP{}
			expectedDesiredLRPSchedulingInfos = []*models.DesiredLRPSchedulingInfo{}
			desiredLRP1 := model_helpers.NewValidDesiredLRP("d-1")
			desiredLRP1.VolumeMounts = nil
			desiredLRP2 := model_helpers.NewValidDesiredLRP("d-2")
			desiredLRP2.VolumeMounts = nil

			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP1)
			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP2)
			for i, expectedDesiredLRP := range expectedDesiredLRPs {
				expectedDesiredLRP.Domain = fmt.Sprintf("domain-%d", i+1)
				Expect(sqlDB.DesireLRP(logger, expectedDesiredLRP)).To(Succeed())
				schedulingInfo := expectedDesiredLRP.DesiredLRPSchedulingInfo()
				expectedDesiredLRPSchedulingInfos = append(expectedDesiredLRPSchedulingInfos, &schedulingInfo)
			}
		})

		It("returns all desired lrps scheduling infos", func() {
			desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPSchedulingInfos).To(HaveLen(2))
			Expect(desiredLRPSchedulingInfos).To(ConsistOf(expectedDesiredLRPSchedulingInfos))
		})

		Context("when filtering by domain", func() {
			It("returns a ResourceNotFound error", func() {
				desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(logger, models.DesiredLRPFilter{Domain: "domain-1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(1))
				Expect(desiredLRPSchedulingInfos[0]).To(BeEquivalentTo(expectedDesiredLRPSchedulingInfos[0]))
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				result, err := db.Exec("UPDATE desired_lrps SET routes = ? WHERE process_guid = ?", "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(1))
			})
		})
	})

	Describe("UpdateDesiredLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP
		var update *models.DesiredLRPUpdate

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(logger, expectedDesiredLRP)).To(Succeed())
			instances := int32(1)
			update = &models.DesiredLRPUpdate{
				Instances: &instances,
			}
		})

		It("updates the lrp", func() {
			instances := int32(123)
			routeContent := []byte("{}")
			routes := models.Routes{
				"blah": (*json.RawMessage)(&routeContent),
			}
			annotation := "annotated"
			update = &models.DesiredLRPUpdate{
				Instances:  &instances,
				Routes:     &routes,
				Annotation: &annotation,
			}
			_, err := sqlDB.UpdateDesiredLRP(logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.Instances = instances
			expectedDesiredLRP.Annotation = annotation
			expectedDesiredLRP.Routes = &routes
			expectedDesiredLRP.ModificationTag.Increment()

			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		It("returns the previous instance count", func() {
			instances := int32(20)
			update = &models.DesiredLRPUpdate{
				Instances: &instances,
			}

			previousInstanceCount, err := sqlDB.UpdateDesiredLRP(logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(previousInstanceCount).To(BeEquivalentTo(1))
		})

		It("updates only the fields in the update parameter", func() {
			instances := int32(20)
			update = &models.DesiredLRPUpdate{
				Instances: &instances,
			}
			_, err := sqlDB.UpdateDesiredLRP(logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.Instances = instances
			expectedDesiredLRP.ModificationTag.Increment()

			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		It("updates only the modification tag if update is empty", func() {
			update = &models.DesiredLRPUpdate{}
			_, err := sqlDB.UpdateDesiredLRP(logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.ModificationTag.Increment()
			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		Context("when routes param is invalid", func() {
			It("returns a bad request error", func() {
				routeContent := []byte("bad json")
				routes := models.Routes{
					"blah": (*json.RawMessage)(&routeContent),
				}
				update = &models.DesiredLRPUpdate{
					Routes: &routes,
				}
				_, err := sqlDB.UpdateDesiredLRP(logger, expectedDesiredLRP.ProcessGuid, update)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrBadRequest))
			})
		})

		Context("when the desired lrp does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				_, err := sqlDB.UpdateDesiredLRP(logger, "does-not-exist", update)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(logger, expectedDesiredLRP)).To(Succeed())
		})

		It("removes the lrp", func() {
			err := sqlDB.RemoveDesiredLRP(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			_, err = sqlDB.DesiredLRPByProcessGuid(logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})

		Context("when the desired lrp does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				err := sqlDB.RemoveDesiredLRP(logger, "does-not-exist")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
