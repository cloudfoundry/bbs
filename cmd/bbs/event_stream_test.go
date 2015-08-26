package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events API", func() {
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

	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

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

		etcdHelper.SetRawDesiredLRP(primerLRP)

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

		err = client.RemoveDesiredLRP(primerLRP.ProcessGuid)
		Expect(err).NotTo(HaveOccurred())

		var event models.Event
		for {
			Eventually(eventChannel).Should(Receive(&event))
			if event.EventType() == models.EventTypeDesiredLRPRemoved {
				break
			}
		}
	})

	AfterEach(func() {
		err := eventSource.Close()
		Expect(err).NotTo(HaveOccurred())
		Eventually(done).Should(BeClosed())
		ginkgomon.Kill(bbsProcess)
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
})
