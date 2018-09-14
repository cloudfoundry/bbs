package metrics

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	logging "code.cloudfoundry.org/diego-logging-client"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
)

const (
	TasksStartedMetric   = "TasksStarted"
	TasksSucceededMetric = "TasksSucceeded"
	TasksFailedMetric    = "TasksFailed"

	PendingTasksMetric   = "TasksPending"
	RunningTasksMetric   = "TasksRunning"
	CompletedTasksMetric = "TasksCompleted"
	ResolvingTasksMetric = "TasksResolving"
)

type perCellStats struct {
	tasksStarted, tasksFailed, tasksSucceeded int
}

type globalStats struct {
	pendingTasks, runningTasks, completedTasks, resolvingTasks int
}

//go:generate counterfeiter -o fakes/fake_task_stat_metron_notifier.go . TaskStatMetronNotifier
type TaskStatMetronNotifier interface {
	ifrit.Runner
	TaskSucceeded(cellID string)
	TaskFailed(cellID string)
	TaskStarted(cellID string)

	TaskConvergenceResults(pending, running, completed, resolved int)
	TaskConvergenceStarted()
}

type taskStatMetronNotifier struct {
	clock                clock.Clock
	mutex                sync.Mutex
	metronClient         logging.IngressClient
	perCellStats         map[string]perCellStats
	globalTaskStats      globalStats
	convergenceRunsDelta uint64
}

func NewTaskStatMetronNotifier(logger lager.Logger, clock clock.Clock, metronClient logging.IngressClient) TaskStatMetronNotifier {
	return &taskStatMetronNotifier{
		clock:        clock,
		perCellStats: make(map[string]perCellStats),
		metronClient: metronClient,
	}
}

func (t *taskStatMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ticker := t.clock.NewTicker(60 * time.Second)

	close(ready)
	for {
		select {
		case <-ticker.C():
			t.emitMetrics()
		case <-signals:
			return nil
		}
	}
}

func (t *taskStatMetronNotifier) emitMetrics() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// emit the metrics
	for cell, stats := range t.perCellStats {
		opt := loggregator.WithEnvelopeTag("cell-id", cell)
		t.metronClient.SendMetric(TasksStartedMetric, stats.tasksStarted, opt)
		t.metronClient.SendMetric(TasksFailedMetric, stats.tasksFailed, opt)
		t.metronClient.SendMetric(TasksSucceededMetric, stats.tasksSucceeded, opt)
	}

	t.metronClient.SendMetric(PendingTasksMetric, t.globalTaskStats.pendingTasks)
	t.metronClient.SendMetric(RunningTasksMetric, t.globalTaskStats.runningTasks)
	t.metronClient.SendMetric(CompletedTasksMetric, t.globalTaskStats.completedTasks)
	t.metronClient.SendMetric(ResolvingTasksMetric, t.globalTaskStats.resolvingTasks)

	t.metronClient.IncrementCounterWithDelta("ConvergenceTaskRuns", t.convergenceRunsDelta)
	println("Resetting")
	t.convergenceRunsDelta = 0
}

func (t *taskStatMetronNotifier) TaskSucceeded(cellID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	stats := t.perCellStats[cellID]
	stats.tasksSucceeded += 1
	t.perCellStats[cellID] = stats
}

func (t *taskStatMetronNotifier) TaskFailed(cellID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	stats := t.perCellStats[cellID]
	stats.tasksFailed += 1
	t.perCellStats[cellID] = stats
}

func (t *taskStatMetronNotifier) TaskStarted(cellID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	stats := t.perCellStats[cellID]
	stats.tasksStarted += 1
	t.perCellStats[cellID] = stats
}

func (t *taskStatMetronNotifier) TaskConvergenceResults(pending, running, completed, resolving int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	println("recording results")

	t.globalTaskStats.pendingTasks = pending
	t.globalTaskStats.runningTasks = running
	t.globalTaskStats.completedTasks = completed
	t.globalTaskStats.resolvingTasks = resolving
}

func (t *taskStatMetronNotifier) TaskConvergenceStarted() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	println("incrementing")
	t.convergenceRunsDelta += 1
}
