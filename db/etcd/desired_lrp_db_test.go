package etcd_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/internal/model_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DesiredLRPDB", func() {
	Describe("DesiredLRPs", func() {
		var filter models.DesiredLRPFilter
		var desiredLRPsInDomains map[string][]*models.DesiredLRP

		Context("when there are desired LRPs", func() {
			var expectedDesiredLRPs []*models.DesiredLRP

			BeforeEach(func() {
				filter = models.DesiredLRPFilter{}
				expectedDesiredLRPs = []*models.DesiredLRP{}

				desiredLRPsInDomains = etcdHelper.CreateDesiredLRPsInDomains(map[string]int{
					"domain-1": 1,
					"domain-2": 2,
				})
			})

			It("returns all the desired LRPs", func() {
				for _, domainLRPs := range desiredLRPsInDomains {
					for _, lrp := range domainLRPs {
						expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
					}
				}
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})

			It("can filter by domain", func() {
				for _, lrp := range desiredLRPsInDomains["domain-2"] {
					expectedDesiredLRPs = append(expectedDesiredLRPs, lrp)
				}
				filter.Domain = "domain-2"
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).To(ConsistOf(expectedDesiredLRPs))
			})
		})

		Context("when there are no LRPs", func() {
			It("returns an empty list", func() {
				desiredLRPs, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(desiredLRPs).NotTo(BeNil())
				Expect(desiredLRPs).To(BeEmpty())
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateValidDesiredLRP("some-guid")
				etcdHelper.CreateMalformedDesiredLRP("some-other-guid")
				etcdHelper.CreateValidDesiredLRP("some-third-guid")
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPs(logger, filter)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DesiredLRPByProcessGuid", func() {
		Context("when there is a desired lrp", func() {
			var desiredLRP *models.DesiredLRP

			BeforeEach(func() {
				desiredLRP = model_helpers.NewValidDesiredLRP("process-guid")
				etcdHelper.SetRawDesiredLRP(desiredLRP)
			})

			It("returns the desired lrp", func() {
				lrp, err := etcdDB.DesiredLRPByProcessGuid(logger, "process-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(lrp).To(Equal(desiredLRP))
			})
		})

		Context("when there is no LRP", func() {
			It("returns a ResourceNotFound", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "nota-guid")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})

		Context("when there is invalid data", func() {
			BeforeEach(func() {
				etcdHelper.CreateMalformedDesiredLRP("some-other-guid")
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrDeserializeJSON))
			})
		})

		Context("when etcd is not there", func() {
			BeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("errors", func() {
				_, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-other-guid")
				Expect(err).To(Equal(models.ErrUnknownError))
			})
		})
	})

	Describe("DesireLRP", func() {
		var lrp *models.DesiredLRP

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
		})

		Context("when the desired LRP does not yet exist", func() {
			It("creates /v1/desired/<process-guid>", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				persisted, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(persisted).To(Equal(lrp))
			})

			It("creates one ActualLRP per index", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
				actualLRPGroups, err := etcdDB.ActualLRPGroupsByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(actualLRPGroups).To(HaveLen(5))
			})

			It("sets a ModificationTag on each ActualLRP with a unique epoch", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
				actualLRPGroups, err := etcdDB.ActualLRPGroupsByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				epochs := map[string]models.ActualLRP{}
				for _, actualLRPGroup := range actualLRPGroups {
					epochs[actualLRPGroup.Instance.ModificationTag.Epoch] = *actualLRPGroup.Instance
				}

				Expect(epochs).To(HaveLen(5))
			})

			It("sets the ModificationTag on the DesiredLRP", func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				lrp, err := etcdDB.DesiredLRPByProcessGuid(logger, "some-process-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(lrp.ModificationTag.Epoch).NotTo(BeEmpty())
				Expect(lrp.ModificationTag.Index).To(BeEquivalentTo(0))
			})

			Context("when an auctioneer is present", func() {
				It("emits start auction requests", func() {
					originalAuctionCallCount := auctioneerClient.RequestLRPAuctionsCallCount()

					err := etcdDB.DesireLRP(logger, lrp)
					Expect(err).NotTo(HaveOccurred())

					desired, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())

					Consistently(auctioneerClient.RequestLRPAuctionsCallCount).Should(Equal(originalAuctionCallCount + 1))

					startAuctions := auctioneerClient.RequestLRPAuctionsArgsForCall(originalAuctionCallCount)
					Expect(startAuctions).To(HaveLen(1))
					Expect(startAuctions[0].DesiredLRP).To(Equal(desired))
					Expect(startAuctions[0].Indices).To(ConsistOf([]uint{0, 1, 2, 3, 4}))
				})
			})
		})

		Context("when the desired LRP already exists", func() {
			var newLRP *models.DesiredLRP

			BeforeEach(func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())

				newLRP = lrp
				newLRP.Instances = 3
			})

			It("rejects the request with ErrResourceExists", func() {
				err := etcdDB.DesireLRP(logger, newLRP)
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("RemoveDesiredLRP", func() {
		var lrp *models.DesiredLRP

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
		})

		Context("when the desired LRP exists", func() {
			BeforeEach(func() {
				err := etcdDB.DesireLRP(logger, lrp)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete it", func() {
				err := etcdDB.RemoveDesiredLRP(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				_, err = etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			Context("when there are running instances on a present cell", func() {
				cellPresence := models.NewCellPresence("the-cell-id", "cell.example.com", "az1", models.NewCellCapacity(128, 1024, 6), []string{}, []string{})

				BeforeEach(func() {
					consulHelper.RegisterCell(cellPresence)

					for i := int32(0); i < lrp.Instances; i++ {
						instanceKey := models.NewActualLRPInstanceKey(fmt.Sprintf("some-instance-guid-%d", i), cellPresence.CellID)
						err := etcdDB.ClaimActualLRP(logger, lrp.ProcessGuid, i, &instanceKey)
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("stops all actual lrps for the desired lrp", func() {
					originalStopCallCount := cellClient.StopLRPInstanceCallCount()

					err := etcdDB.RemoveDesiredLRP(logger, lrp.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())

					Expect(cellClient.StopLRPInstanceCallCount()).To(Equal(originalStopCallCount + int(lrp.Instances)))

					stoppedActuals := make([]int32, lrp.Instances)
					for i := int32(0); i < lrp.Instances; i++ {
						addr, key, _ := cellClient.StopLRPInstanceArgsForCall(originalStopCallCount + int(i))
						Expect(addr).To(Equal(cellPresence.RepAddress))

						stoppedActuals[i] = key.Index
					}

					Expect(stoppedActuals).To(ConsistOf([]int32{0, 1, 2, 3, 4}))
				})
			})
		})

		Context("when the desired LRP does not exist", func() {
			It("returns an resource not found", func() {
				err := etcdDB.RemoveDesiredLRP(logger, "monkey")
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("Updating DesireLRP", func() {
		var (
			update     *models.DesiredLRPUpdate
			desiredLRP *models.DesiredLRP
			lrp        *models.DesiredLRP
		)

		BeforeEach(func() {
			lrp = model_helpers.NewValidDesiredLRP("some-process-guid")
			lrp.Instances = 5
			err := etcdDB.DesireLRP(logger, lrp)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err = etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			update = &models.DesiredLRPUpdate{}
		})

		Context("When the updates are valid", func() {
			BeforeEach(func() {
				annotation := "new-annotation"
				instances := int32(16)

				rawMessage := json.RawMessage([]byte(`{"port":8080,"hosts":["new-route-1","new-route-2"]}`))
				update.Routes = &models.Routes{
					"router": &rawMessage,
				}
				update.Annotation = &annotation
				update.Instances = &instances
			})

			It("updates an existing DesireLRP", func() {
				modelErr := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
				Expect(modelErr).NotTo(HaveOccurred())

				updated, modelErr := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(modelErr).NotTo(HaveOccurred())

				Expect(*updated.Routes).To(HaveKey("router"))
				json, err := (*update.Routes)["router"].MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				updatedJson, err := (*updated.Routes)["router"].MarshalJSON()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedJson).To(MatchJSON(string(json)))
				Expect(updated.Annotation).To(Equal(*update.Annotation))
				Expect(updated.Instances).To(Equal(*update.Instances))
				Expect(updated.ModificationTag.Epoch).To(Equal(desiredLRP.ModificationTag.Epoch))
				Expect(updated.ModificationTag.Index).To(Equal(desiredLRP.ModificationTag.Index + 1))
			})

			Context("when the instances are increased", func() {
				BeforeEach(func() {
					instances := int32(6)
					update.Instances = &instances
				})

				Context("when an auctioneer is present", func() {
					It("emits start auction requests", func() {
						originalAuctionCallCount := auctioneerClient.RequestLRPAuctionsCallCount()

						err := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(err).NotTo(HaveOccurred())

						Consistently(auctioneerClient.RequestLRPAuctionsCallCount).Should(Equal(originalAuctionCallCount + 1))

						updated, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
						Expect(err).NotTo(HaveOccurred())

						startAuctions := auctioneerClient.RequestLRPAuctionsArgsForCall(originalAuctionCallCount)
						Expect(startAuctions).To(HaveLen(1))
						Expect(startAuctions[0].DesiredLRP).To(Equal(updated))
						Expect(startAuctions[0].Indices).To(HaveLen(1))
						Expect(startAuctions[0].Indices).To(ContainElement(uint(5)))
					})
				})
			})

			Context("when the instances are decreased", func() {
				BeforeEach(func() {
					instances := int32(2)
					update.Instances = &instances
				})

				Context("when the cell is present", func() {
					cellPresence := models.NewCellPresence("the-cell-id", "cell.example.com", "az1", models.NewCellCapacity(128, 1024, 6), []string{}, []string{})

					BeforeEach(func() {
						consulHelper.RegisterCell(cellPresence)

						for i := int32(0); i < lrp.Instances; i++ {
							instanceKey := models.NewActualLRPInstanceKey(fmt.Sprintf("some-instance-guid-%d", i), cellPresence.CellID)
							err := etcdDB.ClaimActualLRP(logger, lrp.ProcessGuid, i, &instanceKey)
							Expect(err).NotTo(HaveOccurred())
						}
					})

					It("stops the instances at the removed indices", func() {
						originalStopCallCount := cellClient.StopLRPInstanceCallCount()

						err := etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
						Expect(err).NotTo(HaveOccurred())

						Expect(cellClient.StopLRPInstanceCallCount()).To(Equal(originalStopCallCount + int(lrp.Instances-*(update.Instances))))

						stoppedActuals := make([]int32, lrp.Instances-*update.Instances)
						for i := int32(0); i < (lrp.Instances - *update.Instances); i++ {
							addr, key, _ := cellClient.StopLRPInstanceArgsForCall(originalStopCallCount + int(i))
							Expect(addr).To(Equal(cellPresence.RepAddress))

							stoppedActuals[i] = key.Index
						}

						Expect(stoppedActuals).To(ConsistOf([]int32{2, 3, 4}))
					})
				})
			})
		})

		Context("When the updates are invalid", func() {
			It("instances cannot be less than zero", func() {
				instances := int32(-1)

				update := &models.DesiredLRPUpdate{
					Instances: &instances,
				}

				desiredBeforeUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				err = etcdDB.UpdateDesiredLRP(logger, lrp.ProcessGuid, update)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("instances"))

				desiredAfterUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				Expect(desiredAfterUpdate).To(Equal(desiredBeforeUpdate))
			})
		})

		Context("When the LRP does not exist", func() {
			It("returns an ErrorKeyNotFound", func() {
				instances := int32(0)

				err := etcdDB.UpdateDesiredLRP(logger, "garbage-guid", &models.DesiredLRPUpdate{
					Instances: &instances,
				})
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
