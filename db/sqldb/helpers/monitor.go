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
	QueriesInFlight() int64
}

type queryMonitor struct {
	durationLock     *sync.RWMutex
	queriesTotal     int64
	queriesSucceeded int64
	queriesFailed    int64
	queriesInFlight  int64
	queryDurationMax time.Duration
}

func NewQueryMonitor() QueryMonitor {
	return &queryMonitor{
		durationLock: &sync.RWMutex{},
	}
}

func (q *queryMonitor) MonitorQuery(f func() error) error {
	atomic.AddInt64(&q.queriesInFlight, 1)
	defer atomic.AddInt64(&q.queriesInFlight, -1)

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

func (q *queryMonitor) QueriesInFlight() int64 {
	return atomic.LoadInt64(&q.queriesInFlight)
}

func (q *queryMonitor) QueryDurationMax() time.Duration {
	var durationMax time.Duration
	q.durationLock.RLock()
	durationMax = q.queryDurationMax
	q.durationLock.RUnlock()

	return durationMax
}

func (q *queryMonitor) ReadAndResetQueriesTotal() int64 {
	return atomic.SwapInt64(&q.queriesTotal, 0)
}

func (q *queryMonitor) ReadAndResetQueriesSucceeded() int64 {
	return atomic.SwapInt64(&q.queriesSucceeded, 0)
}

func (q *queryMonitor) ReadAndResetQueriesFailed() int64 {
	return atomic.SwapInt64(&q.queriesFailed, 0)
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
