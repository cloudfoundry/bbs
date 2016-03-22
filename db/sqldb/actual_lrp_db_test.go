package sqldb_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActualLRPDB", func() {
	BeforeEach(func() {
		fakeGUIDProvider.NextGUIDReturns("my-awesome-guid", nil)
	})

	Describe("CreateUnclaimedActualLRP", func() {
		var key *models.ActualLRPKey

		BeforeEach(func() {
			key = &models.ActualLRPKey{
				ProcessGuid: "the-guid",
				Index:       0,
				Domain:      "the-domain",
			}
		})

		It("persists the actual lrp into the database", func() {
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, key)).To(Succeed())

			actualLRP := models.NewUnclaimedActualLRP(*key, fakeClock.Now().Truncate(time.Microsecond).UnixNano())
			actualLRP.ModificationTag.Epoch = "my-awesome-guid"
			actualLRP.ModificationTag.Index = 0

			group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, key.ProcessGuid, key.Index)
			Expect(err).NotTo(HaveOccurred())
			Expect(group).NotTo(BeNil())
			Expect(group.Instance).To(BeEquivalentTo(actualLRP))
			Expect(group.Evacuating).To(BeNil())
		})

		Context("when generating a guid fails", func() {
			BeforeEach(func() {
				fakeGUIDProvider.NextGUIDReturns("", errors.New("no guid for you"))
			})

			It("returns the error", func() {
				err := sqlDB.CreateUnclaimedActualLRP(logger, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrGUIDGeneration))
			})
		})

		Context("when the actual lrp already exists", func() {
			BeforeEach(func() {
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, key)).To(Succeed())
			})

			It("returns a ResourceExists error", func() {
				err := sqlDB.CreateUnclaimedActualLRP(logger, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceExists))
			})
		})
	})

	Describe("ActualLRPGroupByProcessGuidAndIndex", func() {
		var actualLRP *models.ActualLRP
		var now int64

		BeforeEach(func() {
			now = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

			actualLRP = &models.ActualLRP{
				ActualLRPKey: models.NewActualLRPKey("some-guid", 0, "some-domain"),
				State:        models.ActualLRPStateUnclaimed,
				ModificationTag: models.ModificationTag{
					Epoch: "my-awesome-guid",
					Index: 0,
				},
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
		})

		It("returns the existing actual lrp group", func() {
			actualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

			group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
			Expect(err).NotTo(HaveOccurred())
			Expect(group).NotTo(BeNil())
			Expect(group.Instance).To(BeEquivalentTo(actualLRP))
			Expect(group.Evacuating).To(BeNil())
		})

		Context("when there's just an evacuating LRP", func() {
			BeforeEach(func() {
				_, err := db.Exec("UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ? AND instance_index = ? AND evacuating = ?", true, actualLRP.ProcessGuid, actualLRP.Index, false)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the existing actual lrp group", func() {
				actualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

				group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())
				Expect(group).NotTo(BeNil())
				Expect(group.Instance).To(BeNil())
				Expect(group.Evacuating).To(BeEquivalentTo(actualLRP))
			})
		})

		Context("when there are both instance and evacuating LRPs", func() {
			BeforeEach(func() {
				_, err := db.Exec("UPDATE actual_lrps SET evacuating = true WHERE process_guid = ?", actualLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
			})

			It("returns the existing actual lrp group", func() {
				actualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

				group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())
				Expect(group).NotTo(BeNil())
				Expect(group.Instance).To(BeEquivalentTo(actualLRP))
				Expect(group.Evacuating).To(BeEquivalentTo(actualLRP))
			})
		})

		Context("when the actual LRP does not exist", func() {
			It("returns a resource not found error", func() {
				group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, "nope", 0)
				Expect(err).To(Equal(models.ErrResourceNotFound))
				Expect(group).To(BeNil())
			})
		})
	})

	Describe("ActualLRPGroups", func() {
		var allActualLRPGroups []*models.ActualLRPGroup

		BeforeEach(func() {
			allActualLRPGroups = []*models.ActualLRPGroup{}
			fakeGUIDProvider.NextGUIDReturns("mod-tag-guid", nil)

			actualLRPKey1 := &models.ActualLRPKey{
				ProcessGuid: "guid1",
				Index:       1,
				Domain:      "domain1",
			}
			instanceKey1 := &models.ActualLRPInstanceKey{
				InstanceGuid: "i-guid1",
				CellId:       "cell1",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey1)).To(Succeed())
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey1.ProcessGuid, actualLRPKey1.Index, instanceKey1)).To(Succeed())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey1,
					ActualLRPInstanceKey: *instanceKey1,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
			})

			actualLRPKey2 := &models.ActualLRPKey{
				ProcessGuid: "guid-2",
				Index:       1,
				Domain:      "domain2",
			}
			instanceKey2 := &models.ActualLRPInstanceKey{
				InstanceGuid: "i-guid2",
				CellId:       "cell1",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey2)).To(Succeed())
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey2.ProcessGuid, actualLRPKey2.Index, instanceKey2)).To(Succeed())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey2,
					ActualLRPInstanceKey: *instanceKey2,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
			})

			actualLRPKey3 := &models.ActualLRPKey{
				ProcessGuid: "guid3",
				Index:       1,
				Domain:      "domain1",
			}
			instanceKey3 := &models.ActualLRPInstanceKey{
				InstanceGuid: "i-guid3",
				CellId:       "cell2",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey3)).To(Succeed())
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey3.ProcessGuid, actualLRPKey3.Index, instanceKey3)).To(Succeed())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey3,
					ActualLRPInstanceKey: *instanceKey3,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
			})

			actualLRPKey4 := &models.ActualLRPKey{
				ProcessGuid: "guid4",
				Index:       1,
				Domain:      "domain2",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey4)).To(Succeed())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey: *actualLRPKey4,
					State:        models.ActualLRPStateUnclaimed,
					Since:        fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 0,
					},
				},
			})

			actualLRPKey5 := &models.ActualLRPKey{
				ProcessGuid: "guid5",
				Index:       1,
				Domain:      "domain2",
			}
			instanceKey5 := &models.ActualLRPInstanceKey{
				InstanceGuid: "i-guid5",
				CellId:       "cell2",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey5)).To(Succeed())
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey5.ProcessGuid, actualLRPKey5.Index, instanceKey5)).To(Succeed())
			_, err := db.Exec("UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ? AND instance_index = ? AND evacuating = ?", true, actualLRPKey5.ProcessGuid, actualLRPKey5.Index, false)
			Expect(err).NotTo(HaveOccurred())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Evacuating: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey5,
					ActualLRPInstanceKey: *instanceKey5,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
			})

			actualLRPKey6 := &models.ActualLRPKey{
				ProcessGuid: "guid6",
				Index:       1,
				Domain:      "domain1",
			}
			instanceKey6 := &models.ActualLRPInstanceKey{
				InstanceGuid: "i-guid6",
				CellId:       "cell2",
			}
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey6)).To(Succeed())
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey6.ProcessGuid, actualLRPKey6.Index, instanceKey6)).To(Succeed())
			_, err = db.Exec("UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ? AND instance_index = ? AND evacuating = ?", true, actualLRPKey6.ProcessGuid, actualLRPKey6.Index, false)

			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey6)).To(Succeed())
			Expect(sqlDB.ClaimActualLRP(logger, actualLRPKey6.ProcessGuid, actualLRPKey6.Index, instanceKey6)).To(Succeed())

			Expect(err).NotTo(HaveOccurred())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey6,
					ActualLRPInstanceKey: *instanceKey6,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
				Evacuating: &models.ActualLRP{
					ActualLRPKey:         *actualLRPKey6,
					ActualLRPInstanceKey: *instanceKey6,
					State:                models.ActualLRPStateClaimed,
					Since:                fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 1,
					},
				},
			})
		})

		It("returns all the actual lrp groups", func() {
			actualLRPGroups, err := sqlDB.ActualLRPGroups(logger, models.ActualLRPFilter{})
			Expect(err).NotTo(HaveOccurred())

			Expect(actualLRPGroups).To(ConsistOf(allActualLRPGroups))
		})

		Context("when filtering on domains", func() {
			It("returns the actual lrp groups in the domain", func() {
				filter := models.ActualLRPFilter{
					Domain: "domain2",
				}
				actualLRPGroups, err := sqlDB.ActualLRPGroups(logger, filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPGroups).To(HaveLen(3))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[1]))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[3]))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[4]))
			})
		})

		Context("when filtering on cell", func() {
			It("returns the actual lrp groups claimed by the cell", func() {
				filter := models.ActualLRPFilter{
					CellID: "cell1",
				}
				actualLRPGroups, err := sqlDB.ActualLRPGroups(logger, filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPGroups).To(HaveLen(2))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[0]))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[1]))
			})
		})

		Context("when filtering on domain and cell", func() {
			It("returns the actual lrp groups in the domain and claimed by the cell", func() {
				filter := models.ActualLRPFilter{
					Domain: "domain1",
					CellID: "cell2",
				}
				actualLRPGroups, err := sqlDB.ActualLRPGroups(logger, filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPGroups).To(HaveLen(2))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[2]))
				Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[5]))
			})
		})
	})

	Describe("ActualLRPGroupsByProcessGuid", func() {
		var allActualLRPGroups []*models.ActualLRPGroup

		BeforeEach(func() {
			allActualLRPGroups = []*models.ActualLRPGroup{}
			fakeGUIDProvider.NextGUIDReturns("mod-tag-guid", nil)

			actualLRPKey1 := &models.ActualLRPKey{
				ProcessGuid: "guid1",
				Index:       0,
				Domain:      "domain1",
			}
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey1)).To(Succeed())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey: *actualLRPKey1,
					State:        models.ActualLRPStateUnclaimed,
					Since:        fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 0,
					},
				},
			})

			actualLRPKey2 := &models.ActualLRPKey{
				ProcessGuid: "guid1",
				Index:       1,
				Domain:      "domain1",
			}
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey2)).To(Succeed())
			_, err := db.Exec("UPDATE actual_lrps SET evacuating = ? WHERE process_guid = ? AND instance_index = ? AND evacuating = ?", true, actualLRPKey2.ProcessGuid, actualLRPKey2.Index, false)

			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey2)).To(Succeed())

			Expect(err).NotTo(HaveOccurred())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey: *actualLRPKey2,
					State:        models.ActualLRPStateUnclaimed,
					Since:        fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 0,
					},
				},
				Evacuating: &models.ActualLRP{
					ActualLRPKey: *actualLRPKey2,
					State:        models.ActualLRPStateUnclaimed,
					Since:        fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 0,
					},
				},
			})

			actualLRPKey3 := &models.ActualLRPKey{
				ProcessGuid: "guid2",
				Index:       0,
				Domain:      "domain1",
			}
			fakeClock.Increment(time.Hour)
			Expect(sqlDB.CreateUnclaimedActualLRP(logger, actualLRPKey3)).To(Succeed())
			Expect(err).NotTo(HaveOccurred())
			allActualLRPGroups = append(allActualLRPGroups, &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey: *actualLRPKey3,
					State:        models.ActualLRPStateClaimed,
					Since:        fakeClock.Now().Truncate(time.Microsecond).UnixNano(),
					ModificationTag: models.ModificationTag{
						Epoch: "mod-tag-guid",
						Index: 0,
					},
				},
			})
		})

		It("returns all the actual lrp groups for the chosen process guid", func() {
			actualLRPGroups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, "guid1")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualLRPGroups).To(HaveLen(2))
			Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[0]))
			Expect(actualLRPGroups).To(ContainElement(allActualLRPGroups[1]))
		})

		Context("when no actual lrps exist for the process guid", func() {
			It("returns an empty slice", func() {
				actualLRPGroups, err := sqlDB.ActualLRPGroupsByProcessGuid(logger, "guid3")
				Expect(err).NotTo(HaveOccurred())

				Expect(actualLRPGroups).To(HaveLen(0))
			})
		})
	})

	Describe("ClaimActualLRP", func() {
		var instanceKey *models.ActualLRPInstanceKey

		BeforeEach(func() {
			instanceKey = &models.ActualLRPInstanceKey{
				InstanceGuid: "the-instance-guid",
				CellId:       "the-cell-id",
			}
		})

		Context("when the actual lrp exists", func() {
			var expectedActualLRP *models.ActualLRP

			BeforeEach(func() {
				expectedActualLRP = &models.ActualLRP{
					ActualLRPKey: models.ActualLRPKey{
						ProcessGuid: "the-guid",
						Index:       1,
						Domain:      "the-domain",
					},
					ModificationTag: models.ModificationTag{
						Epoch: "my-awesome-guid",
						Index: 0,
					},
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &expectedActualLRP.ActualLRPKey)).To(Succeed())
				fakeClock.Increment(time.Hour)
			})

			Context("and the actual lrp is UNCLAIMED", func() {
				It("claims the actual lrp", func() {
					Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())

					expectedActualLRP.State = models.ActualLRPStateClaimed
					expectedActualLRP.ActualLRPInstanceKey = *instanceKey
					expectedActualLRP.ModificationTag.Increment()
					expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					Expect(actualLRPGroup.Evacuating).To(BeNil())
				})

				Context("and there is a placement error", func() {
					BeforeEach(func() {
						_, err := db.Exec(`
								UPDATE actual_lrps SET placement_error = ?
								WHERE process_guid = ? AND instance_index = ?`,
							"i am placement errror, how are you?",
							expectedActualLRP.ProcessGuid,
							expectedActualLRP.Index,
						)
						Expect(err).NotTo(HaveOccurred())
					})

					It("clears the placement error", func() {
						Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())

						expectedActualLRP.State = models.ActualLRPStateClaimed
						expectedActualLRP.ActualLRPInstanceKey = *instanceKey
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
						Expect(actualLRPGroup.Evacuating).To(BeNil())
					})
				})
			})

			Context("and the actual lrp is CLAIMED", func() {
				Context("when the actual lrp is already claimed with the same instance key", func() {
					BeforeEach(func() {
						Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())
						expectedActualLRP.ModificationTag.Increment()
					})

					It("increments the modification tag", func() {
						Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())

						expectedActualLRP.State = models.ActualLRPStateClaimed
						expectedActualLRP.ActualLRPInstanceKey = *instanceKey
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
						Expect(actualLRPGroup.Evacuating).To(BeNil())
					})
				})

				Context("when the actual lrp is claimed by another cell", func() {
					BeforeEach(func() {
						Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())

						group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(group).NotTo(BeNil())
						Expect(group.Instance).NotTo(BeNil())
						expectedActualLRP = group.Instance
					})

					It("returns an error", func() {
						instanceKey = &models.ActualLRPInstanceKey{
							InstanceGuid: "different-instance",
							CellId:       "different-cell",
						}

						err := sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)
						Expect(err).To(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
						Expect(actualLRPGroup.Evacuating).To(BeNil())
					})
				})
			})

			Context("and the actual lrp is RUNNING", func() {
				BeforeEach(func() {
					netInfo := models.ActualLRPNetInfo{
						Address: "0.0.0.0",
						Ports:   []*models.PortMapping{},
					}

					netInfoData, err := serializer.Marshal(logger, format.ENCODED_PROTO, &netInfo)
					Expect(err).NotTo(HaveOccurred())

					_, err = db.Exec(`
								UPDATE actual_lrps SET state = ?, net_info = ?, cell_id = ?, instance_guid = ?
								WHERE process_guid = ? AND instance_index = ?`,
						models.ActualLRPStateRunning,
						netInfoData,
						instanceKey.CellId,
						instanceKey.InstanceGuid,
						expectedActualLRP.ProcessGuid,
						expectedActualLRP.Index,
					)
					Expect(err).NotTo(HaveOccurred())
				})

				Context("with the same cell and instance guid", func() {
					It("reverts the RUNNING actual lrp to the CLAIMED state", func() {
						Expect(sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)).To(Succeed())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP.ActualLRPInstanceKey = *instanceKey
						expectedActualLRP.State = models.ActualLRPStateClaimed
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
						expectedActualLRP.ModificationTag.Increment()
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
						Expect(actualLRPGroup.Evacuating).To(BeNil())
					})
				})

				Context("with a different cell id", func() {
					BeforeEach(func() {
						instanceKey.CellId = "another-cell"

						group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(group).NotTo(BeNil())
						Expect(group.Instance).NotTo(BeNil())
						expectedActualLRP = group.Instance
					})

					It("returns an error", func() {
						err := sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)
						Expect(err).To(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})

				Context("with a different instance guid", func() {
					BeforeEach(func() {
						instanceKey.InstanceGuid = "another-instance-guid"

						group, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(group).NotTo(BeNil())
						Expect(group.Instance).NotTo(BeNil())
						expectedActualLRP = group.Instance
					})

					It("returns an error", func() {
						err := sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)
						Expect(err).To(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
						Expect(err).NotTo(HaveOccurred())
						Expect(actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})
			})

			Context("and the actual lrp is CRASHED", func() {
				BeforeEach(func() {
					_, err := db.Exec(`
							UPDATE actual_lrps SET state = ?
							WHERE process_guid = ? AND instance_index = ?`,
						models.ActualLRPStateCrashed,
						expectedActualLRP.ProcessGuid,
						expectedActualLRP.Index,
					)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					err := sqlDB.ClaimActualLRP(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index, instanceKey)
					Expect(err).To(HaveOccurred())

					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, expectedActualLRP.ProcessGuid, expectedActualLRP.Index)
					Expect(err).NotTo(HaveOccurred())
					Expect(actualLRPGroup.Instance.State).To(Equal(models.ActualLRPStateCrashed))
				})
			})
		})

		Context("when the actual lrp does not exist", func() {
			BeforeEach(func() {
				key := models.ActualLRPKey{
					ProcessGuid: "the-right-guid",
					Index:       1,
					Domain:      "the-domain",
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &key)).To(Succeed())

				key = models.ActualLRPKey{
					ProcessGuid: "the-wrong-guid",
					Index:       0,
					Domain:      "the-domain",
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &key)).To(Succeed())
			})

			It("returns a ResourceNotFound error", func() {
				err := sqlDB.ClaimActualLRP(logger, "i-do-not-exist", 1, instanceKey)
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("StartActualLRP", func() {
		Context("when the actual lrp exists", func() {
			var (
				instanceKey *models.ActualLRPInstanceKey
				netInfo     *models.ActualLRPNetInfo
				actualLRP   *models.ActualLRP
			)

			BeforeEach(func() {
				instanceKey = &models.ActualLRPInstanceKey{
					InstanceGuid: "the-instance-guid",
					CellId:       "the-cell-id",
				}

				netInfo = &models.ActualLRPNetInfo{
					Address: "1.2.1.2",
					Ports:   []*models.PortMapping{{ContainerPort: 8080, HostPort: 9090}},
				}

				actualLRP = &models.ActualLRP{
					ActualLRPKey: models.ActualLRPKey{
						ProcessGuid: "the-guid",
						Index:       1,
						Domain:      "the-domain",
					},
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
				fakeClock.Increment(time.Hour)
			})

			Context("and the actual lrp is UNCLAIMED", func() {
				It("transitions the state to RUNNING", func() {
					err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
					Expect(err).NotTo(HaveOccurred())

					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
					Expect(err).NotTo(HaveOccurred())

					expectedActualLRP := *actualLRP
					expectedActualLRP.ActualLRPInstanceKey = *instanceKey
					expectedActualLRP.State = models.ActualLRPStateRunning
					expectedActualLRP.ActualLRPNetInfo = *netInfo
					expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
					expectedActualLRP.ModificationTag = models.ModificationTag{
						Epoch: "my-awesome-guid",
						Index: 1,
					}

					Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
				})
			})

			Context("and the actual lrp has been CLAIMED", func() {
				BeforeEach(func() {
					Expect(sqlDB.ClaimActualLRP(logger, actualLRP.ProcessGuid, actualLRP.Index, instanceKey)).To(Succeed())
					fakeClock.Increment(time.Hour)
				})

				Context("and the actual lrp is CLAIMED", func() {
					It("transitions the state to RUNNING", func() {
						err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.ActualLRPInstanceKey = *instanceKey
						expectedActualLRP.State = models.ActualLRPStateRunning
						expectedActualLRP.ActualLRPNetInfo = *netInfo
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
						expectedActualLRP.ModificationTag = models.ModificationTag{
							Epoch: "my-awesome-guid",
							Index: 2,
						}

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})

					Context("and the instance key is different", func() {
						It("transitions the state to RUNNING, updating the instance key", func() {
							otherInstanceKey := &models.ActualLRPInstanceKey{CellId: "some-other-cell", InstanceGuid: "some-other-instance-guid"}
							err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, otherInstanceKey, netInfo)
							Expect(err).NotTo(HaveOccurred())

							actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
							Expect(err).NotTo(HaveOccurred())

							expectedActualLRP := *actualLRP
							expectedActualLRP.ActualLRPInstanceKey = *otherInstanceKey
							expectedActualLRP.State = models.ActualLRPStateRunning
							expectedActualLRP.ActualLRPNetInfo = *netInfo
							expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
							expectedActualLRP.ModificationTag = models.ModificationTag{
								Epoch: "my-awesome-guid",
								Index: 2,
							}

							Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
						})
					})
				})

				Context("and the actual lrp is RUNNING", func() {
					BeforeEach(func() {
						err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
						Expect(err).NotTo(HaveOccurred())
					})

					Context("and the instance key is the same", func() {
						Context("and the net info is the same", func() {
							It("does nothing", func() {
								beforeActualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
								Expect(err).NotTo(HaveOccurred())

								err = sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
								Expect(err).NotTo(HaveOccurred())

								afterActualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
								Expect(err).NotTo(HaveOccurred())
								beforeActualLRPGroup.Instance.ModificationTag.Increment()

								Expect(beforeActualLRPGroup).To(BeEquivalentTo(afterActualLRPGroup))
							})
						})

						Context("and the net info is NOT the same", func() {
							It("updates the net info", func() {
								beforeActualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
								Expect(err).NotTo(HaveOccurred())

								newNetInfo := &models.ActualLRPNetInfo{Address: "some-other-address"}
								err = sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, newNetInfo)
								Expect(err).NotTo(HaveOccurred())

								afterActualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
								Expect(err).NotTo(HaveOccurred())

								beforeActualLRPGroup.Instance.ActualLRPNetInfo = *newNetInfo
								beforeActualLRPGroup.Instance.ModificationTag.Increment()
								Expect(beforeActualLRPGroup).To(BeEquivalentTo(afterActualLRPGroup))
							})
						})
					})

					Context("and the instance key is not the same", func() {
						It("returns an ErrActualLRPCannotBeStarted", func() {
							err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, &models.ActualLRPInstanceKey{CellId: "some-other-cell", InstanceGuid: "some-other-instance-guid"}, netInfo)
							Expect(err).To(Equal(models.ErrActualLRPCannotBeStarted))
						})
					})
				})

				Context("and the actual lrp is CRASHED", func() {
					BeforeEach(func() {
						_, err := db.Exec(`
								UPDATE actual_lrps SET state = ?
								WHERE process_guid = ? AND instance_index = ?`,
							models.ActualLRPStateCrashed,
							actualLRP.ProcessGuid,
							actualLRP.Index,
						)
						Expect(err).NotTo(HaveOccurred())
					})

					It("transitions the state to RUNNING", func() {
						err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.ActualLRPInstanceKey = *instanceKey
						expectedActualLRP.State = models.ActualLRPStateRunning
						expectedActualLRP.ActualLRPNetInfo = *netInfo
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
						expectedActualLRP.ModificationTag = models.ModificationTag{
							Epoch: "my-awesome-guid",
							Index: 2,
						}

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})
			})
		})

		Context("when the actual lrp does not exist", func() {
			var (
				instanceKey *models.ActualLRPInstanceKey
				netInfo     *models.ActualLRPNetInfo
				actualLRP   *models.ActualLRP
			)

			BeforeEach(func() {
				instanceKey = &models.ActualLRPInstanceKey{
					InstanceGuid: "the-instance-guid",
					CellId:       "the-cell-id",
				}

				netInfo = &models.ActualLRPNetInfo{
					Address: "1.2.1.2",
					Ports:   []*models.PortMapping{{ContainerPort: 8080, HostPort: 9090}},
				}

				actualLRP = &models.ActualLRP{
					ActualLRPKey: models.ActualLRPKey{
						ProcessGuid: "the-guid",
						Index:       1,
						Domain:      "the-domain",
					},
					ModificationTag: models.ModificationTag{
						Epoch: "my-awesome-guid",
						Index: 0,
					},
				}
			})

			It("creates the actual lrp", func() {
				err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
				Expect(err).NotTo(HaveOccurred())

				actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())

				expectedActualLRP := *actualLRP
				expectedActualLRP.State = models.ActualLRPStateRunning
				expectedActualLRP.ActualLRPNetInfo = *netInfo
				expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

				Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
			})
		})
	})

	Describe("CrashActualLRP", func() {
		Context("when the actual lrp exists", func() {
			var (
				instanceKey *models.ActualLRPInstanceKey
				netInfo     *models.ActualLRPNetInfo
				actualLRP   *models.ActualLRP
			)

			BeforeEach(func() {
				instanceKey = &models.ActualLRPInstanceKey{
					InstanceGuid: "the-instance-guid",
					CellId:       "the-cell-id",
				}

				netInfo = &models.ActualLRPNetInfo{
					Address: "1.2.1.2",
					Ports:   []*models.PortMapping{{ContainerPort: 8080, HostPort: 9090}},
				}

				actualLRP = &models.ActualLRP{
					ActualLRPKey: models.ActualLRPKey{
						ProcessGuid: "the-guid",
						Index:       1,
						Domain:      "the-domain",
					},
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
				actualLRP.ModificationTag.Epoch = "my-awesome-guid"
				fakeClock.Increment(time.Hour)
			})

			Context("and it is RUNNING", func() {
				BeforeEach(func() {
					err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
					Expect(err).NotTo(HaveOccurred())
					actualLRP.ModificationTag.Increment()
				})

				Context("and it should be restarted", func() {
					It("updates the lrp and sets its state to UNCLAIMED", func() {
						err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.State = models.ActualLRPStateUnclaimed
						expectedActualLRP.CrashCount = 1
						expectedActualLRP.CrashReason = "because it didn't go well"
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})

				Context("and it should NOT be restarted", func() {
					BeforeEach(func() {
						_, err := db.Exec(`
								UPDATE actual_lrps SET crash_count = ?
								WHERE process_guid = ? AND instance_index = ?`,
							models.DefaultImmediateRestarts+2,
							actualLRP.ProcessGuid,
							actualLRP.Index,
						)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the lrp and sets its state to CRASHED", func() {
						err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.State = models.ActualLRPStateCrashed
						expectedActualLRP.CrashCount = models.DefaultImmediateRestarts + 3
						expectedActualLRP.CrashReason = "because it didn't go well"
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})

					Context("and it has NOT been updated recently", func() {
						BeforeEach(func() {
							_, err := db.Exec(`
								UPDATE actual_lrps SET since = ?
								WHERE process_guid = ? AND instance_index = ?`,
								fakeClock.Now().Add(-(models.CrashResetTimeout + 1*time.Second)),
								actualLRP.ProcessGuid,
								actualLRP.Index,
							)
							Expect(err).NotTo(HaveOccurred())
						})

						It("resets the crash count to 1", func() {
							err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
							Expect(err).NotTo(HaveOccurred())

							actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
							Expect(err).NotTo(HaveOccurred())

							Expect(actualLRPGroup.Instance.CrashCount).To(BeNumerically("==", 1))
						})
					})
				})
			})

			Context("and it's CLAIMED", func() {
				BeforeEach(func() {
					Expect(sqlDB.ClaimActualLRP(logger, actualLRP.ProcessGuid, actualLRP.Index, instanceKey)).To(Succeed())
					fakeClock.Increment(time.Hour)
					actualLRP.ModificationTag.Increment()
				})

				Context("and it should be restarted", func() {
					BeforeEach(func() {
						_, err := db.Exec(`
								UPDATE actual_lrps SET crash_count = ?
								WHERE process_guid = ? AND instance_index = ?`,
							models.DefaultImmediateRestarts-1,
							actualLRP.ProcessGuid,
							actualLRP.Index,
						)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the lrp and sets its state to UNCLAIMED", func() {
						err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
						expectedActualLRP.State = models.ActualLRPStateUnclaimed
						expectedActualLRP.CrashCount = models.DefaultImmediateRestarts
						expectedActualLRP.CrashReason = "because it didn't go well"
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})

				Context("and it should NOT be restarted", func() {
					BeforeEach(func() {
						_, err := db.Exec(`
								UPDATE actual_lrps SET crash_count = ?
								WHERE process_guid = ? AND instance_index = ?`,
							models.DefaultImmediateRestarts+2,
							actualLRP.ProcessGuid,
							actualLRP.Index,
						)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the lrp and sets its state to CRASHED", func() {
						err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "some other failure reason")
						Expect(err).NotTo(HaveOccurred())

						actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
						Expect(err).NotTo(HaveOccurred())

						expectedActualLRP := *actualLRP
						expectedActualLRP.State = models.ActualLRPStateCrashed
						expectedActualLRP.CrashCount = models.DefaultImmediateRestarts + 3
						expectedActualLRP.CrashReason = "some other failure reason"
						expectedActualLRP.ModificationTag.Increment()
						expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()

						Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
					})
				})
			})

			Context("and it's already CRASHED", func() {
				BeforeEach(func() {
					err := sqlDB.StartActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, netInfo)
					Expect(err).NotTo(HaveOccurred())
					actualLRP.ModificationTag.Increment()

					err = sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
					Expect(err).NotTo(HaveOccurred())
					actualLRP.ModificationTag.Increment()
				})

				It("returns a cannot crash error", func() {
					err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(models.ErrActualLRPCannotBeCrashed))
				})
			})

			Context("and it's UNCLAIMED", func() {
				It("returns a cannot crash error", func() {
					err := sqlDB.CrashActualLRP(logger, &actualLRP.ActualLRPKey, instanceKey, "because it didn't go well")
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(models.ErrActualLRPCannotBeCrashed))
				})
			})
		})

		Context("when the actual lrp does NOT exist", func() {
			It("returns a record not found error", func() {
				instanceKey := &models.ActualLRPInstanceKey{
					InstanceGuid: "the-instance-guid",
					CellId:       "the-cell-id",
				}

				key := &models.ActualLRPKey{
					ProcessGuid: "the-guid",
					Index:       1,
					Domain:      "the-domain",
				}

				err := sqlDB.CrashActualLRP(logger, key, instanceKey, "because it didn't go well")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("FailActualLRP", func() {
		var actualLRPKey = &models.ActualLRPKey{
			ProcessGuid: "the-guid",
			Index:       1,
			Domain:      "the-domain",
		}

		Context("when the actualLRP exists", func() {
			var actualLRP *models.ActualLRP

			BeforeEach(func() {
				actualLRP = &models.ActualLRP{
					ActualLRPKey: *actualLRPKey,
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
				fakeClock.Increment(time.Hour)
			})

			Context("and the state is UNCLAIMED", func() {
				It("fails the LRP", func() {
					Expect(sqlDB.FailActualLRP(logger, &actualLRP.ActualLRPKey, "failing the LRP")).To(Succeed())

					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
					Expect(err).NotTo(HaveOccurred())

					expectedActualLRP := *actualLRP
					expectedActualLRP.State = models.ActualLRPStateUnclaimed
					expectedActualLRP.PlacementError = "failing the LRP"
					expectedActualLRP.Since = fakeClock.Now().Truncate(time.Microsecond).UnixNano()
					expectedActualLRP.ModificationTag = models.ModificationTag{
						Epoch: "my-awesome-guid",
						Index: 1,
					}

					Expect(*actualLRPGroup.Instance).To(BeEquivalentTo(expectedActualLRP))
				})
			})

			Context("and the state is not UNCLAIMED", func() {
				BeforeEach(func() {
					instanceKey := &models.ActualLRPInstanceKey{
						InstanceGuid: "the-instance-guid",
						CellId:       "the-cell-id",
					}
					Expect(sqlDB.ClaimActualLRP(logger, actualLRP.ProcessGuid, actualLRP.Index, instanceKey)).To(Succeed())
					fakeClock.Increment(time.Hour)
				})

				It("returns a cannot be failed error", func() {
					err := sqlDB.FailActualLRP(logger, &actualLRP.ActualLRPKey, "failing the LRP")
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(models.ErrActualLRPCannotBeFailed))
				})
			})
		})

		Context("when the actualLRP does not exist", func() {
			It("returns a not found error", func() {
				err := sqlDB.FailActualLRP(logger, actualLRPKey, "failing the LRP")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("RemoveActualLRP", func() {
		var actualLRPKey = &models.ActualLRPKey{
			ProcessGuid: "the-guid",
			Index:       1,
			Domain:      "the-domain",
		}

		Context("when the actual LRP exists", func() {
			var actualLRP *models.ActualLRP
			var otherActualLRPKey = &models.ActualLRPKey{
				ProcessGuid: "other-guid",
				Index:       1,
				Domain:      "the-domain",
			}

			BeforeEach(func() {
				actualLRP = &models.ActualLRP{
					ActualLRPKey: *actualLRPKey,
				}
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
				Expect(sqlDB.CreateUnclaimedActualLRP(logger, otherActualLRPKey)).To(Succeed())
				fakeClock.Increment(time.Hour)
			})

			It("removes the actual lrp", func() {
				err := sqlDB.RemoveActualLRP(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())

				_, err = sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})

			It("keeps the other lrps around", func() {
				err := sqlDB.RemoveActualLRP(logger, actualLRP.ProcessGuid, actualLRP.Index)
				Expect(err).NotTo(HaveOccurred())

				_, err = sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, otherActualLRPKey.ProcessGuid, otherActualLRPKey.Index)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the actual lrp does NOT exist", func() {
			It("returns a resource not found error", func() {
				err := sqlDB.RemoveActualLRP(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})

	Describe("UnclaimActualLRP", func() {
		var (
			actualLRP *models.ActualLRP
			guid      = "the-guid"
			index     = int32(1)

			actualLRPKey = &models.ActualLRPKey{
				ProcessGuid: guid,
				Index:       index,
				Domain:      "the-domain",
			}
		)

		Context("when the actual LRP exists", func() {
			Context("When the actual LRP is claimed", func() {
				BeforeEach(func() {
					actualLRP = &models.ActualLRP{
						ActualLRPKey: *actualLRPKey,
					}

					Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
					Expect(sqlDB.ClaimActualLRP(logger, guid, index, &actualLRP.ActualLRPInstanceKey)).To(Succeed())

					_, err := sqlDB.UnclaimActualLRP(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
				})

				It("unclaims the actual LRP", func() {
					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					Expect(actualLRPGroup.Instance.State).To(Equal(models.ActualLRPStateUnclaimed))
				})

				It("it removes the net info from the actualLRP", func() {
					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					Expect(actualLRPGroup.Instance.ActualLRPNetInfo).To(Equal(models.ActualLRPNetInfo{}))
				})

				It("it increments the modification tag on the actualLRP", func() {
					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					// +2 because of claim AND unclaim
					Expect(actualLRPGroup.Instance.ModificationTag.Index).To(Equal(actualLRP.ModificationTag.Index + uint32(2)))
				})

				It("it clears the actualLRP's instance key", func() {
					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					Expect(actualLRPGroup.Instance.ActualLRPInstanceKey).To(Equal(models.ActualLRPInstanceKey{}))
				})

				It("it updates the actualLRP's update at timestamp", func() {
					actualLRPGroup, err := sqlDB.ActualLRPGroupByProcessGuidAndIndex(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					Expect(actualLRPGroup.Instance.Since).To(BeNumerically(">", actualLRP.Since))
				})
			})

			Context("When the actual LRP is unclaimed", func() {
				BeforeEach(func() {
					actualLRP = &models.ActualLRP{
						ActualLRPKey: *actualLRPKey,
					}

					Expect(sqlDB.CreateUnclaimedActualLRP(logger, &actualLRP.ActualLRPKey)).To(Succeed())
				})

				It("stateDidChange is false", func() {
					didStateChange, err := sqlDB.UnclaimActualLRP(logger, guid, index)
					Expect(err).ToNot(HaveOccurred())
					Expect(didStateChange).To(BeFalse())
				})
			})
		})

		Context("when the actual LRP doesn't exist", func() {
			It("returns a resource not found error", func() {
				_, err := sqlDB.UnclaimActualLRP(logger, actualLRPKey.ProcessGuid, actualLRPKey.Index)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrResourceNotFound))
			})
		})
	})
})
