package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	sonde_events "github.com/cloudfoundry/sonde-go/events"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events API", func() {
	var (
		eventChannel chan models.Event

		eventSource    events.EventSource
		baseLRP        *models.ActualLRP
		desiredLRP     *models.DesiredLRP
		key            models.ActualLRPKey
		instanceKey    models.ActualLRPInstanceKey
		newInstanceKey models.ActualLRPInstanceKey
		netInfo        models.ActualLRPNetInfo
	)

	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		Eventually(eventChannel).Should(BeClosed())
		ginkgomon.Kill(bbsProcess)
	})

	Describe("Legacy Events", func() {
		JustBeforeEach(func() {
			var err error
			eventSource, err = client.SubscribeToEvents()
			Expect(err).NotTo(HaveOccurred())

			eventChannel = streamEvents(eventSource)

			primerLRP := model_helpers.NewValidDesiredLRP("primer-guid")
			primeEventStream(eventChannel, models.EventTypeDesiredLRPRemoved, func() {
				etcdHelper.SetRawDesiredLRP(primerLRP)
			}, func() {
				etcdHelper.DeleteDesiredLRP(primerLRP.ProcessGuid)
			})
		})

		It("does not emit latency metrics", func() {
			eventSource.Close()

			timeout := time.After(50 * time.Millisecond)
			for {
				select {
				case envelope := <-testMetricsChan:
					if envelope.GetEventType() == sonde_events.Envelope_ValueMetric {
						Expect(*envelope.ValueMetric.Name).NotTo(Equal("RequestLatency"))
					}
				case <-timeout:
					return
				}
			}
		})

		It("emits request counting metrics", func() {
			eventSource.Close()

			timeout := time.After(50 * time.Millisecond)
			var delta uint64
		OUTER_LOOP:
			for {
				select {
				case envelope := <-testMetricsChan:
					if envelope.GetEventType() == sonde_events.Envelope_CounterEvent {
						counter := envelope.CounterEvent
						if *counter.Name == "RequestCount" {
							delta = *counter.Delta
							break OUTER_LOOP
						}
					}
				case <-timeout:
					break OUTER_LOOP
				}
			}

			Expect(delta).To(BeEquivalentTo(1))
		})
	})

	Describe("Actual LRPs", func() {
		const (
			processGuid     = "some-process-guid"
			domain          = "some-domain"
			noExpirationTTL = 0
		)

		BeforeEach(func() {
			key = models.NewActualLRPKey(processGuid, 0, domain)
			instanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
			newInstanceKey = models.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.1.1.1")

			desiredLRP = &models.DesiredLRP{
				ProcessGuid: processGuid,
				Domain:      domain,
				RootFs:      "some:rootfs",
				Instances:   1,
				Action: models.WrapAction(&models.RunAction{
					Path: "true",
					User: "me",
				}),
			}

			baseLRP = &models.ActualLRP{
				ActualLRPKey: key,
				State:        models.ActualLRPStateUnclaimed,
				Since:        time.Now().UnixNano(),
			}
		})

		JustBeforeEach(func() {
			var err error
			eventSource, err = client.SubscribeToActualLRPEvents()
			Expect(err).NotTo(HaveOccurred())

			eventChannel = streamEvents(eventSource)

			primerLRP := model_helpers.NewValidActualLRP("primer-guid", 0)
			primeEventStream(eventChannel, models.EventTypeActualLRPRemoved, func() {
				etcdHelper.SetRawActualLRP(primerLRP)
			}, func() {
				etcdHelper.DeleteActualLRP(primerLRP.ProcessGuid, primerLRP.Index)
			})
		})

		AfterEach(func() {
			err := eventSource.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("receives events", func() {
			By("creating a ActualLRP")
			etcdHelper.SetRawDesiredLRP(desiredLRP)
			etcdHelper.SetRawActualLRP(baseLRP)

			actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(processGuid, 0)
			Expect(err).NotTo(HaveOccurred())

			var event models.Event
			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPCreatedEvent{}))

			actualLRPCreatedEvent := event.(*models.ActualLRPCreatedEvent)
			Expect(actualLRPCreatedEvent.ActualLrpGroup).To(Equal(actualLRPGroup))

			By("updating the existing ActualLRP")
			err = client.ClaimActualLRP(processGuid, int(key.Index), &instanceKey)
			Expect(err).NotTo(HaveOccurred())

			before := actualLRPGroup
			actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(processGuid, 0)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

			actualLRPChangedEvent := event.(*models.ActualLRPChangedEvent)
			Expect(actualLRPChangedEvent.Before).To(Equal(before))
			Expect(actualLRPChangedEvent.After).To(Equal(actualLRPGroup))

			By("evacuating the ActualLRP")
			_, err = client.EvacuateRunningActualLRP(&key, &instanceKey, &netInfo, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(auctioneerServer.ReceivedRequests()).To(HaveLen(1))
			request := auctioneerServer.ReceivedRequests()[0]
			Expect(request.Method).To(Equal("POST"))
			Expect(request.RequestURI).To(Equal("/v1/lrps"))

			evacuatingLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(processGuid, 0)
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
			err = client.StartActualLRP(&key, &newInstanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())

			// discard instance -> RUNNING
			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

			evacuatingBefore := evacuatingLRP
			_, err = client.EvacuateRunningActualLRP(&key, &newInstanceKey, &netInfo, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(auctioneerServer.ReceivedRequests()).To(HaveLen(2))
			request = auctioneerServer.ReceivedRequests()[1]
			Expect(request.Method).To(Equal("POST"))
			Expect(request.RequestURI).To(Equal("/v1/lrps"))

			evacuatingLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			evacuatingLRP = *evacuatingLRPGroup.GetEvacuating()

			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

			actualLRPChangedEvent = event.(*models.ActualLRPChangedEvent)
			response = actualLRPChangedEvent.Before.GetEvacuating()
			Expect(*response).To(Equal(evacuatingBefore))

			response = actualLRPChangedEvent.After.GetEvacuating()
			Expect(*response).To(Equal(evacuatingLRP))

			// discard instance -> UNCLAIMED
			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPChangedEvent{}))

			By("removing the instance ActualLRP")
			actualLRPGroup, err = client.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP := *actualLRPGroup.GetInstance()

			err = client.RemoveActualLRP(key.ProcessGuid, int(key.Index))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.Event {
				Eventually(eventChannel).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(&models.ActualLRPRemovedEvent{}))

			actualLRPRemovedEvent := event.(*models.ActualLRPRemovedEvent)
			response = actualLRPRemovedEvent.ActualLrpGroup.GetInstance()
			Expect(*response).To(Equal(actualLRP))

			By("removing the evacuating ActualLRP")
			err = client.RemoveEvacuatingActualLRP(&key, &newInstanceKey)
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
			}
		})

		JustBeforeEach(func() {
			var err error
			eventSource, err = client.SubscribeToDesiredLRPEvents()
			Expect(err).NotTo(HaveOccurred())

			eventChannel = streamEvents(eventSource)

			primerLRP := model_helpers.NewValidDesiredLRP("primer-guid")
			primeEventStream(eventChannel, models.EventTypeDesiredLRPRemoved, func() {
				etcdHelper.SetRawDesiredLRP(primerLRP)
			}, func() {
				etcdHelper.DeleteDesiredLRP(primerLRP.ProcessGuid)
			})
		})

		AfterEach(func() {
			err := eventSource.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("receives events", func() {
			By("creating a DesiredLRP")
			err := client.DesireLRP(desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := client.DesiredLRPByProcessGuid(desiredLRP.ProcessGuid)
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
			err = client.UpdateDesiredLRP(desiredLRP.ProcessGuid, &models.DesiredLRPUpdate{Routes: newRoutes})
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))

			desiredLRPChangedEvent, ok := event.(*models.DesiredLRPChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPChangedEvent.After.Routes).To(Equal(newRoutes))

			By("removing the DesiredLRP")
			err = client.RemoveDesiredLRP(desiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))

			desiredLRPRemovedEvent, ok := event.(*models.DesiredLRPRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPRemovedEvent.DesiredLrp.ProcessGuid).To(Equal(desiredLRP.ProcessGuid))
		})
	})

	Describe("Tasks", func() {
		var taskGuid string
		var taskDef *models.TaskDefinition

		BeforeEach(func() {
			taskGuid = "example-guid"
			taskDef = &models.TaskDefinition{
				RootFs: "http://neopets.com",
				Action: models.WrapAction(&models.RunAction{
					User:      "me",
					Dir:       "/tmp",
					Path:      "true",
					LogSource: "logs",
				}),
			}
		})

		JustBeforeEach(func() {
			var err error
			eventSource, err = client.SubscribeToTaskEvents()
			Expect(err).NotTo(HaveOccurred())

			eventChannel = streamEvents(eventSource)
		})

		AfterEach(func() {
			err := eventSource.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("receives events", func() {
			By("creating a Task")
			err := client.DesireTask(taskGuid, "domain", taskDef)
			Expect(err).NotTo(HaveOccurred())

			task, err := client.TaskByGuid(taskGuid)
			Expect(err).NotTo(HaveOccurred())

			var event models.Event
			Eventually(eventChannel).Should(Receive(&event))

			taskCreatedEvent, ok := event.(*models.TaskCreatedEvent)
			Expect(ok).To(BeTrue())

			Expect(taskCreatedEvent.Task).To(Equal(task))

			By("updating an existing Task")
			err = client.FailTask(task.TaskGuid, "i failed")
			Expect(err).NotTo(HaveOccurred())

			Eventually(eventChannel).Should(Receive(&event))

			taskChangedEvent, ok := event.(*models.TaskChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(taskChangedEvent.Before.FailureReason).To(BeEmpty())
			Expect(taskChangedEvent.After.FailureReason).To(Equal("i failed"))

			By("removing the Task")
			etcdHelper.DeleteTask(task.TaskGuid)
			Eventually(eventChannel).Should(Receive(&event))

			taskRemovedEvent, ok := event.(*models.TaskRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(taskRemovedEvent.Task.TaskGuid).To(Equal(task.TaskGuid))
		})
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
