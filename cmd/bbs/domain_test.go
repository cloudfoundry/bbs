package main_test

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/cmd/bbs/testrunner"
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	events "github.com/cloudfoundry/sonde-go/events"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain API", func() {
	BeforeEach(func() {
		bbsRunner = testrunner.New(bbsBinPath, bbsArgs)
		bbsProcess = ginkgomon.Invoke(bbsRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(bbsProcess)
	})

	Describe("UpsertDomain", func() {
		var existingDomain string

		BeforeEach(func() {
			existingDomain = "existing-domain"
			_, err := etcdClient.Set(etcddb.DomainSchemaPath(existingDomain), "", 100)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does emit latency metrics", func() {
			err := client.UpsertDomain(existingDomain, 200*time.Second)
			Expect(err).ToNot(HaveOccurred())

			var sawRequestLatency bool
			timeout := time.After(50 * time.Millisecond)
		OUTER_LOOP:
			for {
				select {
				case envelope := <-testMetricsChan:
					if envelope.GetEventType() == events.Envelope_ValueMetric {
						if *envelope.ValueMetric.Name == "RequestLatency" {
							sawRequestLatency = true
						}
					}
				case <-timeout:
					break OUTER_LOOP
				}
			}
			Expect(sawRequestLatency).To(BeTrue())
		})

		It("emits request counting metrics", func() {
			err := client.UpsertDomain(existingDomain, 200*time.Second)
			Expect(err).ToNot(HaveOccurred())

			timeout := time.After(50 * time.Millisecond)
			var delta uint64
		OUTER_LOOP:
			for {
				select {
				case envelope := <-testMetricsChan:
					if envelope.GetEventType() == events.Envelope_CounterEvent {
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

		It("updates the TTL when updating an existing domain", func() {
			err := client.UpsertDomain(existingDomain, 200*time.Second)
			Expect(err).ToNot(HaveOccurred())

			// etcdEntry, err := etcdClient.Get(etcddb.DomainSchemaPath(existingDomain), false, false)
			// Expect(err).ToNot(HaveOccurred())
			// Expect(etcdEntry.Node.TTL).To(BeNumerically(">", 100))
		})

		It("creates a domain with the desired TTL", func() {
			err := client.UpsertDomain("new-domain", 54*time.Second)
			Expect(err).ToNot(HaveOccurred())

			//etcdEntry, err := etcdClient.Get(etcddb.DomainSchemaPath("new-domain"), false, false)
			//Expect(err).ToNot(HaveOccurred())
			//Expect(etcdEntry.Node.TTL).To(BeNumerically("<=", 54))
		})
	})

	Describe("Domains", func() {
		var expectedDomains []string
		var actualDomains []string
		var getErr error

		BeforeEach(func() {
			expectedDomains = []string{"domain-0", "domain-1"}
			for i, d := range expectedDomains {
				_, err := etcdClient.Set(etcddb.DomainSchemaPath(d), "", uint64(100*(i+1)))
				Expect(err).NotTo(HaveOccurred())
			}

			actualDomains, getErr = client.Domains()
		})

		It("responds without error", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})

		It("has the correct number of responses", func() {
			Expect(actualDomains).To(HaveLen(2))
		})

		It("has the correct domains from the bbs", func() {
			Expect(expectedDomains).To(ConsistOf(actualDomains))
		})
	})
})
