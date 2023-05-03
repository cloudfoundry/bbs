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
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
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
			eventSource, err = client.SubscribeToEvents(logger)
		} else {
			eventSource, err = client.SubscribeToEventsByCellID(logger, cellID)
		}
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Legacy Events", func() {
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
					MetricTags: map[string]*models.MetricTagValue{"some-tag": {Static: "some-value"}},
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
					MetricTags: map[string]*models.MetricTagValue{"some-tag": {Static: "some-value"}},
				}
			})

			Context("Without cell-id filtering", func() {
				It("receives events", func() {
					By("creating a ActualLRP")
					err := client.DesireLRP(logger, "some-trace-id", desiredLRP)
					Expect(err).NotTo(HaveOccurred())

					actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
					Expect(err).NotTo(HaveOccurred())

					var event models.Event
					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

					actualLRPCreatedEvent := event.(*models.ActualLRPCreatedEvent)
					Expect(actualLRPCreatedEvent.ActualLrpGroup).To(Equal(actualLRPGroup))

					By("failing to place the lrp")
					err = client.FailActualLRP(logger, "some-trace-id", &key, "some failure")
					Expect(err).NotTo(HaveOccurred())

					// the lrp group has changed and has a placement error
					actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
					Expect(err).NotTo(HaveOccurred())

					Eventually(eventChannel).Should(Receive(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{})))

					By("updating the existing ActualLRP")
					err = client.ClaimActualLRP(logger, "some-trace-id", &key, &instanceKey)
					Expect(err).NotTo(HaveOccurred())

					before := actualLRPGroup
					actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

					actualLRPChangedEvent := event.(*models.ActualLRPChangedEvent)
					Expect(actualLRPChangedEvent.Before).To(Equal(before))
					Expect(actualLRPChangedEvent.After).To(Equal(actualLRPGroup))

					By("evacuating the ActualLRP")
					initialAuctioneerRequests := auctioneerServer.ReceivedRequests()
					_, err = client.EvacuateRunningActualLRP(logger, "some-trace-id", &key, &instanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{})
					Expect(err).NotTo(HaveOccurred())
					auctioneerRequests := auctioneerServer.ReceivedRequests()
					Expect(auctioneerRequests).To(HaveLen(len(initialAuctioneerRequests) + 1))
					request := auctioneerRequests[len(auctioneerRequests)-1]
					Expect(request.Method).To(Equal("POST"))
					Expect(request.RequestURI).To(Equal("/v1/lrps"))

					evacuatingLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
					Expect(err).NotTo(HaveOccurred())
					evacuatingLRP := *evacuatingLRPGroup.GetEvacuating()

					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

					actualLRPCreatedEvent = event.(*models.ActualLRPCreatedEvent)
					response := actualLRPCreatedEvent.ActualLrpGroup.GetEvacuating()
					Expect(*response).To(Equal(evacuatingLRP))

					// discard instance -> UNCLAIMED
					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

					By("starting and then evacuating the ActualLRP on another cell")
					err = client.StartActualLRP(logger, "some-trace-id", &key, &newInstanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{})

					Expect(err).NotTo(HaveOccurred())

					// discard instance -> RUNNING
					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

					initialAuctioneerRequests = auctioneerServer.ReceivedRequests()
					_, err = client.EvacuateRunningActualLRP(logger, "some-trace-id", &key, &newInstanceKey, &netInfo, []*models.ActualLRPInternalRoute{}, map[string]string{})
					Expect(err).NotTo(HaveOccurred())
					auctioneerRequests = auctioneerServer.ReceivedRequests()
					Expect(auctioneerRequests).To(HaveLen(len(initialAuctioneerRequests) + 1))
					request = auctioneerRequests[len(auctioneerRequests)-1]
					Expect(request.Method).To(Equal("POST"))
					Expect(request.RequestURI).To(Equal("/v1/lrps"))

					evacuatingLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", desiredLRP.ProcessGuid, 0)
					Expect(err).NotTo(HaveOccurred())
					evacuatingLRP = *evacuatingLRPGroup.GetEvacuating()

					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

					actualLRPCreatedEvent = event.(*models.ActualLRPCreatedEvent)
					response = actualLRPCreatedEvent.ActualLrpGroup.GetEvacuating()
					Expect(*response).To(Equal(evacuatingLRP))

					// discard instance -> UNCLAIMED
					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

					By("removing the instance ActualLRP")
					actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", desiredLRP.ProcessGuid, 0)
					Expect(err).NotTo(HaveOccurred())
					actualLRP := *actualLRPGroup.GetInstance()

					err = client.RemoveActualLRP(logger, "some-trace-id", &key, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPRemovedEvent{}))

					actualLRPRemovedEvent := event.(*models.ActualLRPRemovedEvent)
					response = actualLRPRemovedEvent.ActualLrpGroup.GetInstance()
					Expect(*response).To(Equal(actualLRP))

					By("removing the evacuating ActualLRP")
					err = client.RemoveEvacuatingActualLRP(logger, "some-trace-id", &key, &newInstanceKey)
					Expect(err).NotTo(HaveOccurred())

					Eventually(func() models.Event {
						Eventually(eventChannel).Should(Receive(&event))
						return event
					}).Should(BeAssignableToTypeOf(&models.ActualLRPRemovedEvent{}))

					actualLRPRemovedEvent = event.(*models.ActualLRPRemovedEvent)
					response = actualLRPRemovedEvent.ActualLrpGroup.GetEvacuating()
					Expect(*response).To(Equal(evacuatingLRP))
				})
			})

			Context("With cell-id filtering", func() {
				var (
					actualLRPGroup *models.ActualLRPGroup
					err            error
					event          models.Event
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

					actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
					Expect(err).NotTo(HaveOccurred())
					Eventually(eventChannel).ShouldNot(Receive())
				})

				Context("when subscribed to events for a spcific cell", func() {
					It("receives only events from the filtered cell", func() {
						claimLRP := func() {
							before, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
							Expect(err).NotTo(HaveOccurred())

							By("claiming the ActualLRP")
							err = client.ClaimActualLRP(logger, "some-trace-id", &key, &instanceKey)
							Expect(err).NotTo(HaveOccurred())

							actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(logger, "some-trace-id", processGuid, 0)
							Expect(err).NotTo(HaveOccurred())

							Eventually(eventChannel).Should(Receive(Equal(&models.ActualLRPChangedEvent{
								Before: before,
								After:  actualLRPGroup,
							})))

							Expect(actualLRPGroup.GetInstance().GetCellId()).To(Equal(cellID))
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
							BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}),
						))

						claimLRP()

						By("removing the instance ActualLRP")
						err = client.RemoveActualLRP(logger, "some-trace-id", &key, &instanceKey)
						Expect(err).NotTo(HaveOccurred())
						Eventually(eventChannel).Should(Receive(&event))

						actualLRPRemovedEvent := event.(*models.ActualLRPRemovedEvent)
						response := actualLRPRemovedEvent.ActualLrpGroup.GetInstance()
						Expect(response).To(Equal(actualLRPGroup.GetInstance()))
						Expect(actualLRPGroup.GetInstance().GetCellId()).To(Equal(cellID))
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
		eventSource, err = client.SubscribeToEvents(logger)
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
