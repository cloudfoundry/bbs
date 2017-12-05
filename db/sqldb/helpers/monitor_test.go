package helpers_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
)

var _ = Describe("QueryMonitor", func() {
	Describe("MonitorQuery", func() {
		It("increases the total queries count by 1", func() {
			q := helpers.NewQueryMonitor()

			q.MonitorQuery(func() error {
				return nil
			})

			Expect(q.QueriesTotal()).To(BeEquivalentTo(1))
			Expect(q.QueriesSucceeded()).To(BeEquivalentTo(1))
			Expect(q.QueriesFailed()).To(BeEquivalentTo(0))
		})

		It("increments the queries failed count on a bad query", func() {
			q := helpers.NewQueryMonitor()

			q.MonitorQuery(func() error {
				return errors.New("boom!")
			})

			Expect(q.QueriesTotal()).To(BeEquivalentTo(1))
			Expect(q.QueriesSucceeded()).To(BeEquivalentTo(0))
			Expect(q.QueriesFailed()).To(BeEquivalentTo(1))
		})

		It("increments the in-flight count for a query in progress", func() {
			q := helpers.NewQueryMonitor()

			blockCh := make(chan struct{})
			go q.MonitorQuery(func() error {
				<-blockCh
				return nil
			})

			defer func() {
				close(blockCh)
				Eventually(q.QueriesInFlight).Should(BeEquivalentTo(0))
			}()

			Eventually(q.QueriesInFlight).Should(BeEquivalentTo(1))
		})

		It("executes queries and updates the maximum time", func() {
			q := helpers.NewQueryMonitor()
			q.MonitorQuery(func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})

			Expect(q.ReadAndResetQueryDurationMax()).To(BeNumerically(">", 0))
		})

		It("doesn't cause any race conditions", func() {
			q := helpers.NewQueryMonitor()

			blockCh := make(chan struct{})

			updateFunc := func() error {
				<-blockCh
				return nil
			}

			go func() {
				q.MonitorQuery(updateFunc)
			}()
			// increase the chance of race condition happening
			go func() {
				q.MonitorQuery(updateFunc)
			}()

			Consistently(q.ReadAndResetQueryDurationMax()).Should(BeNumerically("==", 0))
			close(blockCh)
			Eventually(q.ReadAndResetQueryDurationMax).Should(BeNumerically(">", 0))
		})
	})

	Describe("ReadAndResetQueryDurationMax", func() {
		It("resets queryDurationMax", func() {
			q := helpers.NewQueryMonitor()
			err := q.MonitorQuery(func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			expectedDuration := q.ReadAndResetQueryDurationMax()
			Expect(expectedDuration).To(BeNumerically(">", 0))
			Expect(q.ReadAndResetQueryDurationMax()).To(BeZero())
		})
	})
})
