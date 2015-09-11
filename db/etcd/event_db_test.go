package etcd_test

import (
	"encoding/json"

	. "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Watchers", func() {
	Describe("WatchForDesiredLRPChanges", func() {
		var (
			creates chan *models.DesiredLRP
			changes chan *models.DesiredLRPChange
			deletes chan *models.DesiredLRP
			stop    chan<- bool
			errors  <-chan error
			lrp     *models.DesiredLRP
		)

		BeforeEach(func() {
			routePayload := json.RawMessage(`{"port":8080,"hosts":["route-1","route-2"]}`)
			lrp = &models.DesiredLRP{
				ProcessGuid: "some-process-guid",
				Domain:      "tests",
				RootFs:      "some:rootfs",
				Action: models.WrapAction(&models.DownloadAction{
					From: "http://example.com",
					To:   "/tmp/internet",
					User: "diego",
				}),
				Routes: &models.Routes{"router": &routePayload},
			}

			creates = make(chan *models.DesiredLRP)
			changes = make(chan *models.DesiredLRPChange)
			deletes = make(chan *models.DesiredLRP)

			stop, errors = etcdDB.WatchForDesiredLRPChanges(logger,
				func(created *models.DesiredLRP) { creates <- created },
				func(changed *models.DesiredLRPChange) { changes <- changed },
				func(deleted *models.DesiredLRP) { deletes <- deleted },
			)
		})

		AfterEach(func() {
			close(stop)
			Consistently(errors).ShouldNot(Receive())
			Eventually(errors).Should(BeClosed())
		})

		It("sends an event down the pipe for creates", func() {
			etcdHelper.SetRawDesiredLRP(lrp)

			desiredLRP, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.GetProcessGuid())
			Expect(err).NotTo(HaveOccurred())

			Eventually(creates).Should(Receive(Equal(desiredLRP)))
		})

		It("sends an event down the pipe for updates", func() {
			etcdHelper.SetRawDesiredLRP(lrp)

			Eventually(creates).Should(Receive())

			desiredBeforeUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.GetProcessGuid())
			Expect(err).NotTo(HaveOccurred())

			lrp.Instances = lrp.GetInstances() + 1
			etcdHelper.SetRawDesiredLRP(lrp)
			Expect(err).NotTo(HaveOccurred())

			desiredAfterUpdate, err := etcdDB.DesiredLRPByProcessGuid(logger, lrp.GetProcessGuid())
			Expect(err).NotTo(HaveOccurred())

			Eventually(changes).Should(Receive(Equal(&models.DesiredLRPChange{
				Before: desiredBeforeUpdate,
				After:  desiredAfterUpdate,
			})))
		})

		It("sends an event down the pipe for deletes", func() {
			etcdHelper.SetRawDesiredLRP(lrp)

			Eventually(creates).Should(Receive())

			desired, bbsErr := etcdDB.DesiredLRPByProcessGuid(logger, lrp.GetProcessGuid())
			Expect(bbsErr).NotTo(HaveOccurred())

			_, err := storeClient.Delete(DesiredLRPSchemaPath(desired), true)
			Expect(err).NotTo(HaveOccurred())

			Eventually(deletes).Should(Receive(Equal(desired)))

		})
	})

	Describe("WatchForActualLRPChanges", func() {
		const (
			lrpProcessGuid = "some-process-guid"
			lrpDomain      = "lrp-domain"
			lrpIndex       = 0
			lrpCellId      = "cell-id"
		)

		var (
			creates chan *models.ActualLRPGroup
			changes chan *models.ActualLRPChange
			deletes chan *models.ActualLRPGroup
			stop    chan<- bool
			errors  <-chan error

			actualLRPGroup *models.ActualLRPGroup
		)

		BeforeEach(func() {
			createsCh := make(chan *models.ActualLRPGroup)
			creates = createsCh

			changesCh := make(chan *models.ActualLRPChange)
			changes = changesCh

			deletesCh := make(chan *models.ActualLRPGroup)
			deletes = deletesCh

			stop, errors = etcdDB.WatchForActualLRPChanges(logger,
				func(created *models.ActualLRPGroup) { createsCh <- created },
				func(changed *models.ActualLRPChange) { changesCh <- changed },
				func(deleted *models.ActualLRPGroup) { deletesCh <- deleted },
			)

			actualLRP := models.NewUnclaimedActualLRP(models.NewActualLRPKey(lrpProcessGuid, lrpIndex, lrpDomain), clock.Now().UnixNano())
			actualLRPGroup = models.NewRunningActualLRPGroup(actualLRP)
		})

		AfterEach(func() {
			close(stop)
			Consistently(errors).ShouldNot(Receive())
			Eventually(errors).Should(BeClosed())
		})

		It("sends an event down the pipe for create", func() {
			etcdHelper.SetRawActualLRP(actualLRPGroup.Instance)
			Eventually(creates).Should(Receive(Equal(actualLRPGroup)))
		})

		It("sends an event down the pipe for updates", func() {
			etcdHelper.SetRawActualLRP(actualLRPGroup.Instance)
			Eventually(creates).Should(Receive())

			updatedGroup := &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         models.NewActualLRPKey(lrpProcessGuid, lrpIndex, lrpDomain),
					ActualLRPInstanceKey: models.NewActualLRPInstanceKey("instance-guid", lrpCellId),
					State:                models.ActualLRPStateClaimed,
					Since:                clock.Now().UnixNano(),
				},
			}
			etcdHelper.SetRawActualLRP(updatedGroup.Instance)

			var actualLRPChange *models.ActualLRPChange
			Eventually(changes).Should(Receive(&actualLRPChange))

			before, after := actualLRPChange.Before, actualLRPChange.After
			Expect(before).To(Equal(actualLRPGroup))
			Expect(after).To(Equal(updatedGroup))
		})

		It("sends an event down the pipe for delete", func() {
			etcdHelper.SetRawActualLRP(actualLRPGroup.Instance)
			Eventually(creates).Should(Receive())

			key := actualLRPGroup.Instance.ActualLRPKey
			_, err := storeClient.Delete(ActualLRPSchemaPath(key.GetProcessGuid(), key.GetIndex()), true)
			Expect(err).NotTo(HaveOccurred())

			Eventually(deletes).Should(Receive(Equal(actualLRPGroup)))
		})

		It("ignores delete events for directories", func() {
			etcdHelper.SetRawActualLRP(actualLRPGroup.Instance)
			Eventually(creates).Should(Receive())

			key := actualLRPGroup.Instance.ActualLRPKey
			_, err := storeClient.Delete(ActualLRPSchemaPath(key.GetProcessGuid(), key.GetIndex()), true)
			Expect(err).NotTo(HaveOccurred())

			Eventually(deletes).Should(Receive(Equal(actualLRPGroup)))

			_, err = storeClient.Delete(ActualLRPProcessDir(key.GetProcessGuid()), true)

			Consistently(logger).ShouldNot(Say("failed-to-unmarshal"))
		})

		Context("when the evacuating key changes", func() {
			It("indicates passes the correct evacuating flag on event callbacks", func() {
				key := models.NewActualLRPKey(lrpProcessGuid, lrpIndex, lrpDomain)

				instanceKey := models.NewActualLRPInstanceKey("instance-guid", "cell-id")
				netInfo := models.NewActualLRPNetInfo("1.1.1.1")
				evacuatedLRPGroup := &models.ActualLRPGroup{
					Evacuating: &models.ActualLRP{
						ActualLRPKey:         key,
						ActualLRPInstanceKey: instanceKey,
						ActualLRPNetInfo:     netInfo,
						State:                models.ActualLRPStateRunning,
						Since:                clock.Now().UnixNano(),
					},
				}

				etcdHelper.SetRawEvacuatingActualLRP(evacuatedLRPGroup.Evacuating, 0)

				Eventually(creates).Should(Receive(Equal(evacuatedLRPGroup)))

				updatedGroup := &models.ActualLRPGroup{
					Evacuating: &models.ActualLRP{
						ActualLRPKey:         key,
						ActualLRPInstanceKey: instanceKey,
						ActualLRPNetInfo:     models.NewActualLRPNetInfo("2.2.2.2"),
						State:                models.ActualLRPStateRunning,
						Since:                clock.Now().UnixNano(),
					},
				}
				etcdHelper.SetRawEvacuatingActualLRP(updatedGroup.Evacuating, 0)

				var actualLRPChange *models.ActualLRPChange
				Eventually(changes).Should(Receive(&actualLRPChange))

				Expect(actualLRPChange.Before).To(Equal(evacuatedLRPGroup))
				Expect(actualLRPChange.After).To(Equal(updatedGroup))

				_, err := storeClient.Delete(EvacuatingActualLRPSchemaPath(key.GetProcessGuid(), key.GetIndex()), true)
				Expect(err).NotTo(HaveOccurred())

				Eventually(deletes).Should(Receive(Equal(updatedGroup)))
			})
		})
	})
})
