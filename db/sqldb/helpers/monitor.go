package helpers

import (
	"database/sql"
	"sync"
	"sync/atomic"
	"time"
)

//go:generate counterfeiter . QueryMonitor
type QueryMonitor interface {
	MonitorQuery(func() error) error
	QueriesTotal() int64
	QueriesSucceeded() int64
	QueriesFailed() int64
	ReadAndResetQueryDurationMax() time.Duration
	ReadAndResetQueriesInFlightMax() int64
}

type queryMonitor struct {
	queriesTotal       int64
	queriesSucceeded   int64
	queriesFailed      int64
	queriesInFlight    int64
	queriesInFlightMax int64
	inFlightLock       *sync.RWMutex
	queryDurationMax   time.Duration
	durationLock       *sync.RWMutex
}

func NewQueryMonitor() QueryMonitor {
	return &queryMonitor{
		durationLock: &sync.RWMutex{},
		inFlightLock: &sync.RWMutex{},
	}
}

func (q *queryMonitor) MonitorQuery(f func() error) error {
	q.updateQueriesInFlight(1)
	defer q.updateQueriesInFlight(-1)

	start := time.Now()
	err := f()
	duration := time.Since(start)

	if err != nil && err != sql.ErrNoRows {
		if err != sql.ErrTxDone {
			atomic.AddInt64(&q.queriesTotal, 1)
			atomic.AddInt64(&q.queriesFailed, 1)
		}
	} else {
		atomic.AddInt64(&q.queriesTotal, 1)
		atomic.AddInt64(&q.queriesSucceeded, 1)
	}

	q.setDurationMax(duration)
	return err
}

func (q *queryMonitor) QueriesTotal() int64 {
	return atomic.LoadInt64(&q.queriesTotal)
}

func (q *queryMonitor) QueriesSucceeded() int64 {
	return atomic.LoadInt64(&q.queriesSucceeded)
}

func (q *queryMonitor) QueriesFailed() int64 {
	return atomic.LoadInt64(&q.queriesFailed)
}

func (q *queryMonitor) ReadAndResetQueriesInFlightMax() int64 {
	var oldMax int64
	q.inFlightLock.Lock()
	oldMax = q.queriesInFlightMax
	q.queriesInFlightMax = q.queriesInFlight
	q.inFlightLock.Unlock()
	return oldMax
}

func (q *queryMonitor) updateQueriesInFlight(delta int64) {
	q.inFlightLock.Lock()
	q.queriesInFlight += delta
	if q.queriesInFlight > q.queriesInFlightMax {
		q.queriesInFlightMax = q.queriesInFlight
	}
	q.inFlightLock.Unlock()
}

func (q *queryMonitor) ReadAndResetQueryDurationMax() time.Duration {
	var oldDuration time.Duration
	q.durationLock.Lock()
	oldDuration = q.queryDurationMax
	q.queryDurationMax = 0
	q.durationLock.Unlock()

	return oldDuration
}

func (q *queryMonitor) setDurationMax(d time.Duration) {
	q.durationLock.Lock()
	if d > q.queryDurationMax {
		q.queryDurationMax = d
	}
	q.durationLock.Unlock()
}
