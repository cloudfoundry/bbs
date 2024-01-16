package sqldb_test

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/bbs/test_helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRPDB", func() {
	Describe("DesireLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP("the-guid")
		})

		It("saves the lrp in the database", func() {
			err := sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, "the-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRP).To(Equal(expectedDesiredLRP))
		})

		Context("when the process_guid is already taken", func() {
			BeforeEach(func() {
				err := sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a resource exists error", func() {
				err := sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("DesiredLRPByProcessGuid", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
		})

		It("returns the desired lrp", func() {
			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		Context("when there are duplicate ports", func() {
			BeforeEach(func() {
				desiredLRPGuid := "desired-lrp-guid-with-duplicate-ports"
				expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
				expectedDesiredLRP.Ports = []uint32{8080, 8080}
				Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
			})

			It("de-dups the ports", func() {
				expectedDesiredLRP.Ports = []uint32{8080}
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRP).To(Equal(expectedDesiredLRP))
			})
		})

		Context("when the desired lrp does not exist", func() {
			It("does not log an error", func() {
				sqlDB.DesiredLRPByProcessGuid(ctx, logger, "Sup dawg")
				Expect(logger.Errors).To(BeEmpty())
			})

			It("returns a ResourceNotFound error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, "Sup dawg")
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(desiredLRP).To(BeNil())
			})
		})

		Context("when the run info is invalid", func() {
			BeforeEach(func() {

				queryStr := `UPDATE desired_lrps SET run_info = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("returns an invalid record error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(desiredLRP).To(BeNil())
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				queryStr := `UPDATE desired_lrps SET routes = ? WHERE process_guid = ?`
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("returns an invalid record error", func() {
				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
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
			expectedDesiredLRPs = append(expectedDesiredLRPs, model_helpers.NewValidDesiredLRP("d-3"))
			for i, expectedDesiredLRP := range expectedDesiredLRPs {
				expectedDesiredLRP.Domain = fmt.Sprintf("domain-%d", i+1)
				Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
			}
		})

		Context("when there are duplicate ports", func() {
			var (
				expectedDesiredLRP *models.DesiredLRP
			)

			BeforeEach(func() {
				desiredLRPGuid := "desired-lrp-guid-with-duplicate-ports"
				expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
				expectedDesiredLRP.Ports = []uint32{8080, 8080}
				Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
			})

			It("de-dups the ports", func() {
				expectedDesiredLRP.Ports = []uint32{8080}
				desiredLRP, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRP).To(ContainElement(expectedDesiredLRP))
			})
		})

		It("returns all desired lrps", func() {
			desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPs).To(HaveLen(3))
			Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
		})

		It("prunes all desired lrps with invalid run infos", func() {
			desiredLRPWithInvalidRunInfo := model_helpers.NewValidDesiredLRP("invalid")
			Expect(sqlDB.DesireLRP(ctx, logger, desiredLRPWithInvalidRunInfo)).To(Succeed())

			queryStr := `UPDATE desired_lrps SET run_info = 'garbage' WHERE process_guid = 'invalid'`
			if test_helpers.UsePostgres() {
				queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
			}
			_, err := db.ExecContext(ctx, queryStr)
			Expect(err).NotTo(HaveOccurred())

			desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPs).To(HaveLen(3))

			rows, err := db.QueryContext(ctx, `SELECT process_guid FROM desired_lrps`)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()

			processGuids := []string{}
			for rows.Next() {
				var processGuid string
				err := rows.Scan(&processGuid)
				Expect(err).NotTo(HaveOccurred())
				processGuids = append(processGuids, processGuid)
			}
			Expect(processGuids).NotTo(ContainElement("invalid"))
		})

		Context("when filtering by domain", func() {
			It("returns the filtered desired lrps", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{Domain: "domain-1"})
				Expect(err).NotTo(HaveOccurred())

				Expect(desiredLRPs).To(HaveLen(1))
				Expect(desiredLRPs[0]).To(BeEquivalentTo(expectedDesiredLRPs[0]))
			})
		})

		Context("when filtering by process guids", func() {
			It("returns the filtered desired lrps", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{ProcessGuids: []string{"d-1", "d-3"}})
				Expect(err).NotTo(HaveOccurred())

				Expect(desiredLRPs).To(HaveLen(2))
				Expect(desiredLRPs).To(ContainElement(expectedDesiredLRPs[0]))
				Expect(desiredLRPs).To(ContainElement(expectedDesiredLRPs[2]))
			})
		})

		Context("when the run info is invalid", func() {
			BeforeEach(func() {
				queryStr := "UPDATE desired_lrps SET run_info = ? WHERE process_guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(HaveLen(2))
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				queryStr := "UPDATE desired_lrps SET routes = ? WHERE process_guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPs, err := sqlDB.DesiredLRPs(ctx, logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(HaveLen(2))
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
			desiredLRP2 := model_helpers.NewValidDesiredLRP("d-2")
			desiredLRP3 := model_helpers.NewValidDesiredLRP("d-3")

			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP1)
			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP2)
			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP3)
			for i, expectedDesiredLRP := range expectedDesiredLRPs {
				expectedDesiredLRP.Domain = fmt.Sprintf("domain-%d", i+1)
				Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
				schedulingInfo := expectedDesiredLRP.DesiredLRPSchedulingInfo()
				expectedDesiredLRPSchedulingInfos = append(expectedDesiredLRPSchedulingInfos, &schedulingInfo)
			}
		})

		It("returns all desired lrps scheduling infos", func() {
			desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPSchedulingInfos).To(HaveLen(3))
			Expect(desiredLRPSchedulingInfos).To(ConsistOf(expectedDesiredLRPSchedulingInfos))
		})

		Context("when filtering by domain", func() {
			It("returns the filtered schedulig infos", func() {
				desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{Domain: "domain-1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(1))
				Expect(desiredLRPSchedulingInfos[0]).To(BeEquivalentTo(expectedDesiredLRPSchedulingInfos[0]))
			})
		})

		Context("when filtering by process guids", func() {
			It("returns the filtered schedulig infos", func() {
				filter := models.DesiredLRPFilter{ProcessGuids: []string{"d-1", "d-3"}}
				desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(2))
				Expect(desiredLRPSchedulingInfos).To(ContainElement(expectedDesiredLRPSchedulingInfos[0]))
				Expect(desiredLRPSchedulingInfos).To(ContainElement(expectedDesiredLRPSchedulingInfos[2]))
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				queryStr := "UPDATE desired_lrps SET routes = ? WHERE process_guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPSchedulingInfos, err := sqlDB.DesiredLRPSchedulingInfos(ctx, logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(2))
			})
		})
	})

	Describe("DesiredLRPSchedulingInfoByProcessGuid", func() {
		var expectedDesiredLRPSchedulingInfo models.DesiredLRPSchedulingInfo

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			desiredLRP := model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(ctx, logger, desiredLRP)).To(Succeed())
			expectedDesiredLRPSchedulingInfo = desiredLRP.DesiredLRPSchedulingInfo()
		})

		It("Returns the desired-lrp", func() {
			schedInfo, err := sqlDB.DesiredLRPSchedulingInfoByProcessGuid(ctx, logger, expectedDesiredLRPSchedulingInfo.ProcessGuid)

			Expect(err).NotTo(HaveOccurred())
			Expect(*schedInfo).To(BeEquivalentTo(expectedDesiredLRPSchedulingInfo))
		})

		Context("when the desired lrp does not exist", func() {
			It("does not log an error", func() {
				sqlDB.DesiredLRPSchedulingInfoByProcessGuid(ctx, logger, "I am Batman")
				Expect(logger.Errors).To(BeEmpty())
			})

			It("returns a ResourceNotFound error", func() {
				schedInfo, err := sqlDB.DesiredLRPSchedulingInfoByProcessGuid(ctx, logger, "I am Batman")
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(schedInfo).To(BeNil())
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				queryStr := "UPDATE desired_lrps SET routes = ? WHERE process_guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}

				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRPSchedulingInfo.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("returns an invalid record error", func() {
				schedInfo, err := sqlDB.DesiredLRPSchedulingInfoByProcessGuid(ctx, logger, expectedDesiredLRPSchedulingInfo.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(schedInfo).To(BeNil())
			})
		})
	})

	Describe("DesiredLRPRoutingInfos", func() {
		var expectedDesiredLRPRoutingInfos []*models.DesiredLRP
		var expectedDesiredLRPs []*models.DesiredLRP

		BeforeEach(func() {
			expectedDesiredLRPs = []*models.DesiredLRP{}
			expectedDesiredLRPRoutingInfos = []*models.DesiredLRP{}
			desiredLRP1 := model_helpers.NewValidDesiredLRP("d-1")
			desiredLRP2 := model_helpers.NewValidDesiredLRP("d-2")
			desiredLRP3 := model_helpers.NewValidDesiredLRP("d-3")

			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP1)
			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP2)
			expectedDesiredLRPs = append(expectedDesiredLRPs, desiredLRP3)
			for i, expectedDesiredLRP := range expectedDesiredLRPs {
				expectedDesiredLRP.Domain = fmt.Sprintf("domain-%d", i+1)
				expectedDesiredLRP.Instances = 0
				Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
				routingInfo := expectedDesiredLRP.DesiredLRPRoutingInfo()
				expectedDesiredLRPRoutingInfos = append(expectedDesiredLRPRoutingInfos, &routingInfo)
			}
		})

		It("returns all desired lrps routing infos", func() {
			desiredLRPRoutingInfos, err := sqlDB.DesiredLRPRoutingInfos(ctx, logger, models.DesiredLRPFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(desiredLRPRoutingInfos).To(HaveLen(3))
			Expect(desiredLRPRoutingInfos).To(ConsistOf(expectedDesiredLRPRoutingInfos))
		})

		Context("when filtering by process guids", func() {
			It("returns the filtered routing infos", func() {
				filter := models.DesiredLRPFilter{ProcessGuids: []string{"d-1", "d-3"}}
				desiredLRPRoutingInfos, err := sqlDB.DesiredLRPRoutingInfos(ctx, logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPRoutingInfos).To(HaveLen(2))
				Expect(desiredLRPRoutingInfos).To(ContainElement(expectedDesiredLRPRoutingInfos[0]))
				Expect(desiredLRPRoutingInfos).To(ContainElement(expectedDesiredLRPRoutingInfos[2]))
			})
		})

		Context("when the routes are invalid", func() {
			BeforeEach(func() {
				queryStr := "UPDATE desired_lrps SET routes = ? WHERE process_guid = ?"
				if test_helpers.UsePostgres() {
					queryStr = test_helpers.ReplaceQuestionMarks(queryStr)
				}
				result, err := db.ExecContext(ctx, queryStr, "{{", expectedDesiredLRPs[0].ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeEquivalentTo(1))
			})

			It("excludes the invalid desired LRP from the response", func() {
				desiredLRPRoutingInfos, err := sqlDB.DesiredLRPRoutingInfos(ctx, logger, models.DesiredLRPFilter{})
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPRoutingInfos).To(HaveLen(2))
			})
		})
	})

	Describe("UpdateDesiredLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP
		var update *models.DesiredLRPUpdate

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
		})

		JustBeforeEach(func() {
			Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
			update = &models.DesiredLRPUpdate{}
			update.SetInstances(1)
		})

		It("updates the lrp", func() {
			routeContent := []byte("{}")
			routes := models.Routes{
				"blah": (*json.RawMessage)(&routeContent),
			}
			update = &models.DesiredLRPUpdate{Routes: &routes}
			update.SetInstances(123)
			update.SetAnnotation("annotated")
			_, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.Instances = 123
			expectedDesiredLRP.Annotation = "annotated"
			expectedDesiredLRP.Routes = &routes
			expectedDesiredLRP.ModificationTag.Increment()

			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		It("returns the desired lrp from before the update", func() {
			update = &models.DesiredLRPUpdate{}
			update.SetInstances(20)

			beforeDesiredLRP, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(beforeDesiredLRP).To(Equal(expectedDesiredLRP))
		})

		It("updates only the fields in the update parameter", func() {
			update = &models.DesiredLRPUpdate{}
			update.SetInstances(20)
			_, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.Instances = 20
			expectedDesiredLRP.ModificationTag.Increment()

			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		It("updates only the modification tag if update is empty", func() {
			update = &models.DesiredLRPUpdate{}
			_, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			expectedDesiredLRP.ModificationTag.Increment()
			Expect(desiredLRP).To(BeEquivalentTo(expectedDesiredLRP))
		})

		Context("when updating the metric tags", func() {
			BeforeEach(func() {
				expectedDesiredLRP.MetricTags["tag-to-be-overwritten"] = &models.MetricTagValue{Static: "overwritten-value"}
			})

			It("overwrites the existing metric tags in the database", func() {
				expectedMetricTags := map[string]*models.MetricTagValue{
					"some-tag": &models.MetricTagValue{Static: "some-value"},
				}
				update = &models.DesiredLRPUpdate{MetricTags: expectedMetricTags}

				_, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
				Expect(err).NotTo(HaveOccurred())

				desiredLRP, err := sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRP.MetricTags).ToNot(HaveKey("tag-to-be-overwritten"))
				Expect(desiredLRP.MetricTags).To(Equal(expectedMetricTags))
			})
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
				_, err := sqlDB.UpdateDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid, update)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrBadRequest))
			})
		})

		Context("when the desired lrp does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				_, err := sqlDB.UpdateDesiredLRP(ctx, logger, "does-not-exist", update)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var expectedDesiredLRP *models.DesiredLRP

		BeforeEach(func() {
			desiredLRPGuid := "desired-lrp-guid"
			expectedDesiredLRP = model_helpers.NewValidDesiredLRP(desiredLRPGuid)
			Expect(sqlDB.DesireLRP(ctx, logger, expectedDesiredLRP)).To(Succeed())
		})

		It("removes the lrp", func() {
			err := sqlDB.RemoveDesiredLRP(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			_, err = sqlDB.DesiredLRPByProcessGuid(ctx, logger, expectedDesiredLRP.ProcessGuid)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrResourceNotFound))
		})

		Context("when the desired lrp does not exist", func() {
			It("returns a ResourceNotFound error", func() {
				err := sqlDB.RemoveDesiredLRP(ctx, logger, "does-not-exist")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
