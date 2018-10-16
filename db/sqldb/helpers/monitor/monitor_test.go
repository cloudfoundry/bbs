package monitor_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers/monitor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitor", func() {
	Describe("Monitor", func() {
		It("increases the total queries count by 1", func() {
			q := monitor.New()

			q.Monitor(func() error {
				return nil
			})

			Expect(q.Total()).To(BeEquivalentTo(1))
			Expect(q.Succeeded()).To(BeEquivalentTo(1))
			Expect(q.Failed()).To(BeEquivalentTo(0))
		})

		It("increments the queries failed count on a bad query", func() {
			q := monitor.New()

			q.Monitor(func() error {
				return errors.New("boom!")
			})

			Expect(q.Total()).To(BeEquivalentTo(1))
			Expect(q.Succeeded()).To(BeEquivalentTo(0))
			Expect(q.Failed()).To(BeEquivalentTo(1))
		})

		It("executes queries and updates the maximum count of queries in flight", func() {
			q := monitor.New()
			startedCh := make(chan struct{})
			blockCh := make(chan struct{})
			go q.Monitor(func() error {
				close(startedCh)
				<-blockCh
				return nil
			})
			defer close(blockCh)

			<-startedCh
			Expect(q.ReadAndResetInFlightMax()).To(Equal(int64(1)))
		})

		It("executes queries and updates the maximum time", func() {
			q := monitor.New()
			q.Monitor(func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})

			Expect(q.ReadAndResetDurationMax()).To(BeNumerically(">", 0))
		})

		It("doesn't cause any race conditions", func() {
			q := monitor.New()

			blockCh := make(chan struct{})

			updateFunc := func() error {
				<-blockCh
				return nil
			}

			go func() {
				q.Monitor(updateFunc)
			}()
			// increase the chance of race condition happening
			go func() {
				q.Monitor(updateFunc)
			}()

			Consistently(q.ReadAndResetDurationMax).Should(BeNumerically("==", 0))
			close(blockCh)
			Eventually(q.ReadAndResetDurationMax).Should(BeNumerically(">", 0))
		})
	})

	Describe("ReadAndResetDurationMax", func() {
		It("resets queryDurationMax", func() {
			q := monitor.New()
			err := q.Monitor(func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			expectedDuration := q.ReadAndResetDurationMax()
			Expect(expectedDuration).To(BeNumerically(">", 0))
			Expect(q.ReadAndResetDurationMax()).To(BeZero())
		})
	})

	Describe("ReadAndResetInFlightMax", func() {
		It("resets queriesInFlightMax", func() {
			q := monitor.New()
			blockCh1 := make(chan struct{})
			startedCh1 := make(chan struct{})
			finishedCh1 := make(chan struct{})
			blockCh2 := make(chan struct{})
			startedCh2 := make(chan struct{})
			finishedCh2 := make(chan struct{})
			go func() {
				q.Monitor(func() error {
					close(startedCh1)
					<-blockCh1
					return nil
				})
				close(finishedCh1)
			}()
			go func() {
				q.Monitor(func() error {
					close(startedCh2)
					<-blockCh2
					return nil
				})
				close(finishedCh2)
			}()

			<-startedCh1
			<-startedCh2

			Consistently(q.ReadAndResetInFlightMax).Should(Equal(int64(2)))
			close(blockCh1)
			<-finishedCh1
			Expect(q.ReadAndResetInFlightMax()).To(Equal(int64(2)))
			Consistently(q.ReadAndResetInFlightMax).Should(Equal(int64(1)))
			close(blockCh2)
			<-finishedCh2
			Expect(q.ReadAndResetInFlightMax()).To(Equal(int64(1)))
			Consistently(q.ReadAndResetInFlightMax).Should(Equal(int64(0)))
		})
	})
})
