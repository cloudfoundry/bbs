package main_test

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/bbs/cmd/bbs/testrunner"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	. "code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events API", func() {
	var (
		eventChannel chan models.Event

		eventSource    events.EventSource
		err            error
		desiredLRP     *models.DesiredLRP
		key            models.ActualLRPKey
		instanceKey    models.ActualLRPInstanceKey
		newInstanceKey models.ActualLRPInstanceKey
		netInfo        models.ActualLRPNetInfo
		cellID         string
	)

	BeforeEach(func() {
		cellID = ""
		bbsRunner = testrunner.New(bbsBinPath, bbsConfig)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	JustBeforeEach(func() {
		if cellID == "" {
			eventSource, err = client.SubscribeToInstanceEvents(logger)
		} else {
			eventSource, err = client.SubscribeToInstanceEventsByCellID(logger, cellID)
		}
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("LRP Events", func() {
		JustBeforeEach(func() {
			eventChannel = streamEvents(eventSource)

			primerLRP := model_helpers.NewValidDesiredLRP("primer-guid")
			primeEventStream(eventChannel, models.EventTypeDesiredLRPRemoved, func() {
				err := client.DesireLRP(logger, "some-trace-id", primerLRP)
				Expect(err).NotTo(HaveOccurred())
			}, func() {
				err := client.RemoveDesiredLRP(logger, "some-trace-id", "primer-guid")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		AfterEach(func() {
			err := eventSource.Close()
			Expect(err).NotTo(HaveOccurred())
			Eventually(eventChannel).Should(BeClosed())
		})

		Describe("Desired LRPs", func() {
			BeforeEach(func() {
				routeMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["original-route"]}]`))
				routes := &models.Routes{"cf-router": &routeMessage}

				desiredLRP = &models.DesiredLRP{
					ProcessGuid: "some-guid",
					Domain:      "some-domain",
					RootFs:      "some:rootfs",
					Routes:      routes,
					Action: models.WrapAction(&models.RunAction{
						User:      "me",
						Dir:       "/tmp",
						Path:      "true",
						LogSource: "logs",
					}),
					MetricTags: &models.MetricTags{"some-tag": {Static: "some-value"}},
				}
			})

			It("receives events", func() {
				By("creating a DesiredLRP")
				err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				desiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", desiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				var event models.Event
				Eventually(eventChannel).Should(Receive(&event))

				desiredLRPCreatedEvent, ok := event.(*models.DesiredLRPCreatedEvent)
				Expect(ok).To(BeTrue())

				Expect(desiredLRPCreatedEvent.DesiredLrp).To(Equal(desiredLRP))

				By("updating an existing DesiredLRP")
				routeMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["new-route"]}]`))
				newRoutes := &models.Routes{
					"cf-router": &routeMessage,
				}
				err = client.UpdateDesiredLRP(logger, "some-trace-id", desiredLRP.ProcessGuid, &models.DesiredLRPUpdate{Routes: newRoutes})
				Expect(err).NotTo(HaveOccurred())

				Eventually(eventChannel).Should(Receive(&event))

				desiredLRPChangedEvent, ok := event.(*models.DesiredLRPChangedEvent)
				Expect(ok).To(BeTrue())
				Expect(desiredLRPChangedEvent.After.Routes).To(Equal(newRoutes))

				By("removing the DesiredLRP")
				err = client.RemoveDesiredLRP(logger, "some-trace-id", desiredLRP.ProcessGuid)
				Expect(err).NotTo(HaveOccurred())

				Eventually(eventChannel).Should(Receive(&event))

				desiredLRPRemovedEvent, ok := event.(*models.DesiredLRPRemovedEvent)
				Expect(ok).To(BeTrue())
				Expect(desiredLRPRemovedEvent.DesiredLrp.ProcessGuid).To(Equal(desiredLRP.ProcessGuid))
			})
		})

		Describe("Actual LRPs", func() {
			const (
				processGuid = "some-process-guid"
				domain      = "some-domain"
			)
			getEvacuatingLRPFromList := func(lrps []*models.ActualLRP) models.ActualLRPInstanceKey {
				for _, lrp := range lrps {
					if lrp.Presence == models.ActualLRP_Evacuating {
						return lrp.ActualLRPInstanceKey
					}
				}
				return models.ActualLRPInstanceKey{}
			}

			BeforeEach(func() {
				key = models.NewActualLRPKey(processGuid, 0, domain)
				instanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
				newInstanceKey = models.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
				netInfo = models.NewActualLRPNetInfo("1.1.1.1", "3.3.3.3", models.ActualLRPNetInfo_PreferredAddressUnknown)

				desiredLRP = &models.DesiredLRP{
					ProcessGuid: processGuid,
					Domain:      domain,
					RootFs:      "some:rootfs",
					Instances:   1,
					Action: models.WrapAction(&models.RunAction{
						Path: "true",
						User: "me",
					}),
					MetricTags: &models.MetricTags{"some-tag": {Static: "some-value"}},
				}
			})

			Context("Without cell-id filtering", func() {
				index0 := int32(0)
				It("receives events", func() {
					By("creating a ActualLRP")
					err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
					Expect(err).NotTo(HaveOccurred())

					actualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index0})
					Expect(err).NotTo(HaveOccurred())

					var created *models.ActualLRPInstanceCreatedEvent
					Eventually(eventChannel).Should(Receive(&created))
					Expect(created.ActualLrp).To(BeEquivalentTo(actualLRPGroup[0]))

					By("failing to place the lrp")
					err = client.FailActualLRP(logger, "some-trace-id", &key, "some failure")
					Expect(err).NotTo(HaveOccurred())

					actualLRPGroup, err = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index0})
					Expect(err).NotTo(HaveOccurred())

					var ce *models.ActualLRPInstanceChangedEvent
					Eventually(eventChannel).Should(Receive(&ce))
					Expect(ce.ActualLRPInstanceKey).To(Equal(actualLRPGroup[0].ActualLRPInstanceKey))
					Expect(ce.Before.State).To(Equal(models.ActualLRPStateUnclaimed))
					Expect(ce.After.State).To(Equal(models.ActualLRPStateUnclaimed))
					Expect(ce.After.PlacementError).ToNot(Equal(""))

					By("updating the existing ActualLRP")
					err = client.ClaimActualLRP(logger, "some-trace-id", &key, &instanceKey)
					Expect(err).NotTo(HaveOccurred())

					actualLRPGroup, err = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index0})
					Expect(err).NotTo(HaveOccurred())

					Eventually(eventChannel).Should(Receive(&ce))
					Expect(ce.ActualLRPInstanceKey).To(Equal(actualLRPGroup[0].ActualLRPInstanceKey))
					Expect(ce.Before.State).To(Equal(models.ActualLRPStateUnclaimed))
					Expect(ce.After.State).To(Equal(models.ActualLRPStateClaimed))

					By("evacuating the ActualLRP")
					initialAuctioneerRequests := auctioneerServer.ReceivedRequests()
					_, err = client.EvacuateRunningActualLRP(logger, "some-trace-id", &key, &instanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{}, false, "")
					Expect(err).NotTo(HaveOccurred())
					auctioneerRequests := auctioneerServer.ReceivedRequests()
					Expect(auctioneerRequests).To(HaveLen(len(initialAuctioneerRequests) + 1))
					request := auctioneerRequests[len(auctioneerRequests)-1]
					Expect(request.Method).To(Equal("POST"))
					Expect(request.RequestURI).To(Equal("/v1/lrps"))

					evacuatingLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index0})
					Expect(err).NotTo(HaveOccurred())

					evacuatingLRP := getEvacuatingLRPFromList(evacuatingLRPGroup)

					Eventually(eventChannel).Should(Receive(&ce))
					Expect(ce.ActualLRPInstanceKey).To(Equal(evacuatingLRP))
					Expect(ce.Before.Presence).To(Equal(models.ActualLRP_Ordinary))
					Expect(ce.After.Presence).To(Equal(models.ActualLRP_Evacuating))

					Eventually(eventChannel).Should(Receive(&created))
					Expect(created.ActualLrp.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))

					By("starting and then evacuating the ActualLRP on another cell")
					err = client.StartActualLRP(logger, "some-trace-id", &key, &newInstanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{}, false, "")
					Expect(err).NotTo(HaveOccurred())

					Eventually(eventChannel).Should(Receive(&ce))
					Expect(ce.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))
					Expect(ce.ActualLRPInstanceKey.CellId).To(Equal("other-cell-id"))
					Expect(ce.Before.State).To(Equal(models.ActualLRPStateUnclaimed))
					Expect(ce.After.State).To(Equal(models.ActualLRPStateRunning))
					Expect(ce.After.Presence).To(Equal(models.ActualLRP_Ordinary))

					initialAuctioneerRequests = auctioneerServer.ReceivedRequests()
					_, err = client.EvacuateRunningActualLRP(logger, "some-trace-id", &key, &newInstanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{}, false, "")
					Expect(err).NotTo(HaveOccurred())
					auctioneerRequests = auctioneerServer.ReceivedRequests()
					Expect(auctioneerRequests).To(HaveLen(len(initialAuctioneerRequests) + 1))
					request = auctioneerRequests[len(auctioneerRequests)-1]
					Expect(request.Method).To(Equal("POST"))
					Expect(request.RequestURI).To(Equal("/v1/lrps"))

					Eventually(eventChannel).Should(Receive(&ce))
					Expect(ce.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))
					Expect(ce.ActualLRPInstanceKey.CellId).To(Equal("other-cell-id"))
					Expect(ce.After.State).To(Equal(models.ActualLRPStateRunning))
					Expect(ce.After.Presence).To(Equal(models.ActualLRP_Evacuating))

					Eventually(eventChannel).Should(Receive(&created))
					Expect(ce.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))

					By("removing the new ActualLRP")
					actualLRPGroup, err = client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: desiredLRP.ProcessGuid, Index: &index0})
					Expect(err).NotTo(HaveOccurred())

					err = client.RemoveActualLRP(logger, "some-trace-id", &key, nil)
					Expect(err).NotTo(HaveOccurred())

					var re *models.ActualLRPInstanceRemovedEvent
					Eventually(eventChannel).Should(Receive(&re))
					Expect(re.ActualLrp.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))
					Expect(re.ActualLrp.ActualLRPInstanceKey.CellId).To(Equal(""))
					Expect(re.ActualLrp.Presence).To(Equal(models.ActualLRP_Ordinary))

					By("removing the evacuating ActualLRP")
					err = client.RemoveEvacuatingActualLRP(logger, "some-trace-id", &key, &newInstanceKey)
					Expect(err).NotTo(HaveOccurred())

					Eventually(eventChannel).Should(Receive(&re))
					Expect(re.ActualLrp.ActualLRPKey).To(Equal(actualLRPGroup[0].ActualLRPKey))
					Expect(re.ActualLrp.ActualLRPInstanceKey.CellId).To(Equal("other-cell-id"))
					Expect(re.ActualLrp.Presence).To(Equal(models.ActualLRP_Evacuating))
				})
			})

			Context("With cell-id filtering", func() {
				var (
					err   error
					event models.Event
				)

				BeforeEach(func() {
					cellID = "cell-id"
				})

				JustBeforeEach(func() {
					err = client.DesireLRP(logger, "some-trace-id", desiredLRP)
					Expect(err).NotTo(HaveOccurred())

					desiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-trace-id", desiredLRP.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())

					Eventually(eventChannel).Should(Receive(&event))

					desiredLRPCreatedEvent, ok := event.(*models.DesiredLRPCreatedEvent)
					Expect(ok).To(BeTrue())

					Expect(desiredLRPCreatedEvent.DesiredLrp).To(Equal(desiredLRP))
					Eventually(eventChannel).ShouldNot(Receive())
				})

				Context("when subscribed to events for a spcific cell", func() {
					It("receives only events from the filtered cell", func() {
						index0 := int32(0)
						claimLRP := func() {
							By("claiming the ActualLRP")
							err = client.ClaimActualLRP(logger, "some-trace-id", &key, &instanceKey)
							Expect(err).NotTo(HaveOccurred())

							actualLRPGroup, err := client.ActualLRPs(logger, "some-trace-id", models.ActualLRPFilter{ProcessGuid: processGuid, Index: &index0})
							Expect(err).NotTo(HaveOccurred())

							var e *models.ActualLRPInstanceChangedEvent
							Eventually(eventChannel).Should(Receive(&e))
							Expect(e.ActualLRPInstanceKey).To(BeEquivalentTo(actualLRPGroup[0].ActualLRPInstanceKey))
							Expect(e.Before.State).To(Equal(models.ActualLRPStateUnclaimed))
							Expect(e.After.State).To(Equal(models.ActualLRPStateClaimed))

							Expect(actualLRPGroup[0].GetCellId()).To(Equal(cellID))
						}
						claimLRP()

						By("crashing the instance ActualLRP")
						var event1, event2 models.Event
						err = client.CrashActualLRP(logger, "some-trace-id", &key, &instanceKey, "booom!!")
						Expect(err).NotTo(HaveOccurred())

						Eventually(eventChannel).Should(Receive(&event1))
						Eventually(eventChannel).Should(Receive(&event2))

						Expect([]models.Event{event1, event2}).To(ConsistOf(
							BeAssignableToTypeOf(&models.ActualLRPCrashedEvent{}),
							BeAssignableToTypeOf(&models.ActualLRPInstanceRemovedEvent{}),
						))

						claimLRP()

						By("removing the instance ActualLRP")
						err = client.RemoveActualLRP(logger, "some-trace-id", &key, &instanceKey)
						Expect(err).NotTo(HaveOccurred())
						Eventually(eventChannel).Should(Receive(&event))

						actualLRPRemovedEvent := event.(*models.ActualLRPInstanceRemovedEvent)
						Expect(actualLRPRemovedEvent.ActualLrp.ActualLRPInstanceKey).To(Equal(instanceKey))
						Expect(actualLRPRemovedEvent.ActualLrp.ActualLRPInstanceKey.CellId).To(Equal(cellID))
					})

					It("does not receive events from the other cells", func() {
						By("updating the existing ActualLRP")
						err = client.ClaimActualLRP(logger, "some-trace-id", &key, &newInstanceKey)
						Expect(err).NotTo(HaveOccurred())

						Consistently(eventChannel).ShouldNot(Receive())

						err = client.RemoveActualLRP(logger, "some-trace-id", &key, &newInstanceKey)
						Expect(err).NotTo(HaveOccurred())
						Consistently(eventChannel).ShouldNot(Receive())
					})
				})
			})
		})

		It("does not emit latency metrics", func() {
			time.Sleep(time.Millisecond * 50)
			eventSource.Close()

			Eventually(testMetricsChan).ShouldNot(Receive(testhelpers.MatchV2Metric(testhelpers.MetricAndValue{
				Name: "RequestLatency",
			})))
		})

		It("emits request counting metrics", func() {
			eventSource.Close()

			var total uint64
			Eventually(testMetricsChan).Should(Receive(
				SatisfyAll(
					WithTransform(func(source *loggregator_v2.Envelope) *loggregator_v2.Counter {
						return source.GetCounter()
					}, Not(BeNil())),
					WithTransform(func(source *loggregator_v2.Envelope) string {
						return source.GetCounter().Name
					}, Equal("RequestCount")),
					WithTransform(func(source *loggregator_v2.Envelope) uint64 {
						total += source.GetCounter().Delta
						return total
					}, BeEquivalentTo(3)),
				),
			))
		})
	})

	Describe("Tasks", func() {
		var (
			taskDef *models.TaskDefinition
		)

		BeforeEach(func() {
			taskDef = model_helpers.NewValidTaskDefinition()
			eventSource, err = client.SubscribeToTaskEvents(logger)
			Expect(err).NotTo(HaveOccurred())
			eventChannel = streamEvents(eventSource)
		})

		It("receives events", func() {
			err := client.DesireTask(logger, "some-trace-id", "completed-task", "some-domain", taskDef)
			Expect(err).NotTo(HaveOccurred())

			task := &models.Task{
				TaskGuid:       "completed-task",
				Domain:         "some-domain",
				TaskDefinition: taskDef,
				State:          models.Task_Pending,
			}

			var event models.Event
			Eventually(eventChannel).Should(Receive(&event))
			taskCreatedEvent, ok := event.(*models.TaskCreatedEvent)
			Expect(ok).To(BeTrue())
			taskCreatedEvent.Task.CreatedAt = 0
			taskCreatedEvent.Task.UpdatedAt = 0
			Expect(taskCreatedEvent.Task).To(DeepEqual(task))

			err = client.CancelTask(logger, "some-trace-id", "completed-task")
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))
			taskChangedEvent, ok := event.(*models.TaskChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(taskChangedEvent.Before.State).To(Equal(models.Task_Pending))
			Expect(taskChangedEvent.After.State).To(Equal(models.Task_Completed))

			err = client.ResolvingTask(logger, "some-trace-id", "completed-task")
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))
			taskChangedEvent, ok = event.(*models.TaskChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(taskChangedEvent.Before.State).To(Equal(models.Task_Completed))
			Expect(taskChangedEvent.After.State).To(Equal(models.Task_Resolving))

			err = client.DeleteTask(logger, "some-trace-id", "completed-task")
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))
			taskRemovedEvent, ok := event.(*models.TaskRemovedEvent)
			Expect(ok).To(BeTrue())

			taskRemovedEvent.Task.CreatedAt = 0
			taskRemovedEvent.Task.UpdatedAt = 0
			taskRemovedEvent.Task.FirstCompletedAt = 0
			task.State = models.Task_Resolving
			task.Failed = true
			task.FailureReason = "task was cancelled"

			Expect(taskRemovedEvent.Task).To(DeepEqual(task))
		})
	})

	It("cleans up exiting connections when killing the BBS", func() {
		var err error
		eventSource, err = client.SubscribeToInstanceEvents(logger)
		Expect(err).NotTo(HaveOccurred())

		done := make(chan struct{})
		go func() {
			_, err := eventSource.Next()
			Expect(err).To(HaveOccurred())
			close(done)
		}()

		taskEventSource, err := client.SubscribeToTaskEvents(logger)
		Expect(err).ToNot(HaveOccurred())

		taskDone := make(chan struct{})
		go func() {
			_, err := taskEventSource.Next()
			Expect(err).To(HaveOccurred())
			close(taskDone)
		}()

		ginkgomon.Interrupt(bbsProcess)
		Eventually(done).Should(BeClosed())
		Eventually(taskDone).Should(BeClosed())
	})
})

func primeEventStream(eventChannel chan models.Event, eventType string, primer func(), cleanup func()) {
	primer()

PRIMING:
	for {
		select {
		case <-eventChannel:
			break PRIMING
		case <-time.After(50 * time.Millisecond):
			cleanup()
			primer()
		}
	}

	cleanup()

	var event models.Event
	for {
		Eventually(eventChannel).Should(Receive(&event))
		if event.EventType() == eventType {
			break
		}
	}

	for {
		select {
		case <-eventChannel:
		case <-time.After(50 * time.Millisecond):
			return
		}
	}
}

func streamEvents(eventSource events.EventSource) chan models.Event {
	eventChannel := make(chan models.Event)

	go func() {
		for {
			event, err := eventSource.Next()
			if err != nil {
				close(eventChannel)
				return
			}
			eventChannel <- event
		}
	}()

	return eventChannel
}
