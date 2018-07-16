package main_test

import (
	"crypto/rand"
	"os"
	"time"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/guidprovider"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/durationjson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Convergence API", func() {
	Describe("ConvergeLRPs", func() {
		var processGuid string

		BeforeEach(func() {
			// make the converger more aggressive by running every second
			bbsConfig.ConvergeRepeatInterval = durationjson.Duration(time.Second)
			bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
			bbsProcess = ginkgomon.Invoke(bbsRunner)

			cellPresence := models.NewCellPresence(
				"some-cell",
				"cell.example.com",
				"http://cell.example.com",
				"the-zone",
				models.NewCellCapacity(128, 1024, 6),
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			consulHelper.RegisterCell(&cellPresence)
			processGuid = "some-process-guid"
			desiredLRP := model_helpers.NewValidDesiredLRP(processGuid)
			err := client.DesireLRP(logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when an LRP cell is dead", func() {
			BeforeEach(func() {
				netInfo := models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.NewPortMapping(8080, 80))

				err := client.StartActualLRP(logger, &models.ActualLRPKey{
					ProcessGuid: processGuid,
					Index:       0,
					Domain:      "some-domain",
				}, &models.ActualLRPInstanceKey{
					InstanceGuid: "ig-1",
					CellId:       "missing-cell",
				}, &netInfo)

				Expect(err).NotTo(HaveOccurred())
			})

			// Row 1 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
			FIt("makes the LRP suspect", func() {
				Eventually(func() models.ActualLRP_Presence {
					group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
					Expect(err).NotTo(HaveOccurred())
					return group.Instance.Presence
				}).Should(Equal(models.ActualLRP_Suspect))
			})

			Context("and the LRP is marked Suspect", func() {
				var (
					db     *sqldb.SQLDB
					events events.EventSource
				)

				BeforeEach(func() {
					var err error
					events, err = client.SubscribeToEvents(logger)
					Expect(err).NotTo(HaveOccurred())

					key, keys, err := bbsConfig.EncryptionConfig.Parse()
					Expect(err).NotTo(HaveOccurred())
					keyManager, err := encryption.NewKeyManager(key, keys)
					cryptor := encryption.NewCryptor(keyManager, rand.Reader)
					wrappedDB := helpers.NewMonitoredDB(sqlRunner.DB(), helpers.NewQueryMonitor())
					metronClient := &testhelpers.FakeIngressClient{}
					db = sqldb.NewSQLDB(
						wrappedDB,
						1,
						1,
						cryptor,
						guidprovider.DefaultGuidProvider,
						clock.NewClock(),
						sqlRunner.DriverName(),
						metronClient,
					)

					Eventually(func() models.ActualLRP_Presence {
						group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
						Expect(err).NotTo(HaveOccurred())
						return group.Instance.Presence
					}).Should(Equal(models.ActualLRP_Suspect))
				})

				// Row 2 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
				Context("when the cell is back", func() {
					BeforeEach(func() {
						cellPresence := models.NewCellPresence(
							"missing-cell",
							"cell.example.com",
							"http://cell.example.com",
							"the-zone",
							models.NewCellCapacity(128, 1024, 6),
							[]string{},
							[]string{},
							[]string{},
							[]string{},
						)
						consulHelper.RegisterCell(&cellPresence)
					})

					FIt("it transitions back to Ordinary", func() {
						Eventually(func() models.ActualLRP_Presence {
							group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
							Expect(err).NotTo(HaveOccurred())
							return group.Instance.Presence
						}).Should(Equal(models.ActualLRP_Ordinary))
					})
				})

				// Row 3 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
				Context("when there is a new Ordinary LRP in Running statea", func() {
					BeforeEach(func() {
						var err error

						netInfo := models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.NewPortMapping(8080, 80))
						_, _, err = db.StartActualLRP(logger, &models.ActualLRPKey{
							ProcessGuid: "some-process-guid",
							Index:       0,
							Domain:      "some-domain",
						}, &models.ActualLRPInstanceKey{
							InstanceGuid: "ig-2",
							CellId:       "some-cell",
						}, &netInfo)
						Expect(err).NotTo(HaveOccurred())
					})

					FIt("removes the suspect LRP", func() {
						var lrp *models.ActualLRP
						Eventually(func() string {
							group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
							Expect(err).NotTo(HaveOccurred())
							lrp = group.Instance
							return lrp.InstanceGuid
						}).Should(Equal("ig-2"))
						Expect(lrp.Presence).To(Equal(models.ActualLRP_Ordinary))
					})

					FIt("emits a LRPRemoved event", func() {
						eventCh := streamEvents(events)
						var removedEvent *models.ActualLRPRemovedEvent
						Eventually(eventCh).Should(Receive(&removedEvent))

						Expect(removedEvent.ActualLrpGroup.Evacuating).To(BeNil())
						Expect(removedEvent.ActualLrpGroup.Instance.InstanceGuid).To(Equal("ig-1"))
						Expect(removedEvent.ActualLrpGroup.Instance.Presence).To(Equal(models.ActualLRP_Suspect))
					})
				})

				// Row 5 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
				Context("when the new Ordinary LRP cells goes missing", func() {
					BeforeEach(func() {
						var err error

						events, err = client.SubscribeToEvents(logger)
						Expect(err).NotTo(HaveOccurred())

						err = client.ClaimActualLRP(logger, &models.ActualLRPKey{
							ProcessGuid: "some-process-guid",
							Index:       0,
							Domain:      "some-domain",
						}, &models.ActualLRPInstanceKey{
							InstanceGuid: "ig-2",
							CellId:       "another-missing-cell",
						})
						Expect(err).NotTo(HaveOccurred())
					})

					FIt("Unclaims the LRP and emits a LRPChanged event", func() {
						eventCh := streamEvents(events)

						var e *models.ActualLRPChangedEvent

						Eventually(func() string {
							Eventually(eventCh).Should(Receive(&e))
							return e.After.Instance.State
						}).Should(Equal(models.ActualLRPStateUnclaimed))

						Expect(e.Before.Instance.InstanceGuid).To(Equal("ig-2"))
						Expect(e.Before.Instance.Presence).To(Equal(models.ActualLRP_Ordinary))
						Expect(e.Before.Instance.State).To(Equal(models.ActualLRPStateClaimed))
						Expect(e.After.Instance.State).To(Equal(models.ActualLRPStateUnclaimed))
					})
				})

				// Row 6 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
				Context("when the Auctioneer calls FailActualLRP", func() {
					BeforeEach(func() {
						err := client.FailActualLRP(logger, &models.ActualLRPKey{
							ProcessGuid: "some-process-guid",
							Index:       0,
							Domain:      "some-domain",
						}, "boom!")
						Expect(err).NotTo(HaveOccurred())
					})

					FIt("keeps the suspect LRP untouched", func() {
						Consistently(func() models.ActualLRP_Presence {
							group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
							Expect(err).NotTo(HaveOccurred())
							return group.Instance.Presence
						}).Should(Equal(models.ActualLRP_Suspect))
					})
				})

				// All tests in this context need to use a non aggressive converger to
				// ensure they are testing state transitions as a result of the RPC
				// calls (.e.g StartActualLRP) instead of testing converger behavior.
				// We also have to initially make the convergence aggressive in the
				// outer Context in order to ensure the LRP transition from Ordinary to
				// Suspect within 1 second instead of waiting for the default 30
				// second.
				Context("with a less aggressive converger", func() {
					BeforeEach(func() {
						bbsProcess.Signal(os.Interrupt)
						Eventually(bbsProcess.Wait()).Should(Receive())
						bbsConfig.ConvergeRepeatInterval = durationjson.Duration(time.Hour)
						bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
						bbsProcess = ginkgomon.Invoke(bbsRunner)

						// recreate the event stream
						var err error
						events, err = client.SubscribeToEvents(logger)
						Expect(err).NotTo(HaveOccurred())
					})

					// Row 7 https://docs.google.com/document/d/19880DjH4nJKzsDP8BT09m28jBlFfSiVx64skbvilbnA/edit
					Context("when the replacement cell is started by calling StartActualLRP", func() {
						BeforeEach(func() {
							netInfo := models.NewActualLRPNetInfo("127.0.0.1", "10.10.10.10", models.NewPortMapping(8080, 80))
							err := client.StartActualLRP(logger, &models.ActualLRPKey{
								ProcessGuid: "some-process-guid",
								Index:       0,
								Domain:      "some-domain",
							}, &models.ActualLRPInstanceKey{
								InstanceGuid: "ig-2",
								CellId:       "some-cell",
							}, &netInfo)
							Expect(err).NotTo(HaveOccurred())
						})

						FIt("replaces the Running LRP instance with the ordinary one", func() {
							group, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, 0)
							Expect(err).NotTo(HaveOccurred())
							Expect(group.Instance.Presence).To(Equal(models.ActualLRP_Ordinary))
						})

						FIt("emits a LRPCreated event", func() {
							eventCh := streamEvents(events)

							var e *models.ActualLRPCreatedEvent

							Eventually(eventCh, 2*time.Second).Should(Receive(&e))
							Expect(e.ActualLrpGroup.Instance.InstanceGuid).To(Equal("ig-2"))
							Expect(e.ActualLrpGroup.Instance.Presence).To(Equal(models.ActualLRP_Ordinary))
						})

						FIt("emits a LRPRemoved event", func() {
							eventCh := streamEvents(events)

							var e *models.ActualLRPRemovedEvent

							Eventually(eventCh, 2*time.Second).Should(Receive(&e))
							Expect(e.ActualLrpGroup.Instance.InstanceGuid).To(Equal("ig-1"))
							Expect(e.ActualLrpGroup.Instance.Presence).To(Equal(models.ActualLRP_Suspect))
						})
					})
				})
			})
		})

		Context("when the lrp goes missing", func() {
			BeforeEach(func() {
				err := client.RemoveActualLRP(logger, &models.ActualLRPKey{
					ProcessGuid: processGuid,
					Index:       0,
					Domain:      "some-domain",
				}, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("converges the lrps", func() {
				Eventually(func() []*models.ActualLRPGroup {
					groups, err := client.ActualLRPGroupsByProcessGuid(logger, processGuid)
					Expect(err).NotTo(HaveOccurred())
					return groups
				}).Should(HaveLen(1))
			})
		})

		Context("when a task is desired but its cell is dead", func() {
			BeforeEach(func() {
				task := model_helpers.NewValidTask("task-guid")

				err := client.DesireTask(logger, task.TaskGuid, task.Domain, task.TaskDefinition)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.StartTask(logger, task.TaskGuid, "dead-cell")
				Expect(err).NotTo(HaveOccurred())
			})

			It("marks the task as completed and failed", func() {
				Eventually(func() []*models.Task {
					return getTasksByState(client, models.Task_Completed)
				}).Should(HaveLen(1))

				Expect(getTasksByState(client, models.Task_Completed)[0].Failed).To(BeTrue())
			})
		})
	})
})

func getTasksByState(client bbs.InternalClient, state models.Task_State) []*models.Task {
	tasks, err := client.Tasks(logger)
	Expect(err).NotTo(HaveOccurred())

	filteredTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.State == state {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks
}
