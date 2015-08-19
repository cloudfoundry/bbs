package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events API", func() {
	Describe("Actual LRPs", func() {
		const (
			processGuid     = "some-process-guid"
			domain          = "some-domain"
			noExpirationTTL = 0
		)

		var (
			done         chan struct{}
			eventChannel chan models.Event

			eventSource    events.EventSource
			baseLRP        *models.ActualLRP
			desiredLRP     *models.DesiredLRP
			key            models.ActualLRPKey
			instanceKey    models.ActualLRPInstanceKey
			newInstanceKey models.ActualLRPInstanceKey
			netInfo        models.ActualLRPNetInfo
		)

		JustBeforeEach(func() {
			var err error
			eventSource, err = client.SubscribeToEvents()
			Expect(err).NotTo(HaveOccurred())

			eventChannel = make(chan models.Event)
			done = make(chan struct{})

			go func() {
				defer close(done)
				for {
					event, err := eventSource.Next()
					if err != nil {
						close(eventChannel)
						return
					}
					eventChannel <- event
				}
			}()

			rawMessage := json.RawMessage([]byte(`{"port":8080,"hosts":["primer-route"]}`))
			primerLRP := &models.DesiredLRP{
				ProcessGuid: "primer-guid",
				Domain:      "primer-domain",
				RootFs:      "primer:rootfs",
				Routes: &models.Routes{
					"router": &rawMessage,
				},
				Action: models.WrapAction(&models.RunAction{
					User: "me",
					Path: "true",
				}),
			}

		PRIMING:
			for {
				select {
				case <-eventChannel:
					break PRIMING
				case <-time.After(50 * time.Millisecond):
					etcdHelper.SetRawDesiredLRP(primerLRP)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})

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

		It("receives events", func() {
			By("creating a ActualLRP")
			// replace me with client.DesiredLRP
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
})
