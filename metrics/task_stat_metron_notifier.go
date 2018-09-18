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
	PrunedTasksMetric    = "ConvergenceTasksPruned"
	KickedTasksMetric    = "ConvergenceTasksKicked"

	ConvergeTaskRunsCounter = "ConvergenceTaskRuns"
	ConvergeTaskDuration    = "ConvergenceTaskDuration"
)

type perCellStats struct {
	tasksStarted, tasksFailed, tasksSucceeded int
}

type globalStats struct {
	pendingTasks, runningTasks, completedTasks, resolvingTasks int
	prunedTasks, kickedTasks                                   uint64
}

//go:generate counterfeiter -o fakes/fake_task_stat_metron_notifier.go . TaskStatMetronNotifier
type TaskStatMetronNotifier interface {
	ifrit.Runner
	TaskSucceeded(cellID string)
	TaskFailed(cellID string)
	TaskStarted(cellID string)

	SnapshotTasks(pending, running, completed, resolved int, pruned, kicked uint64)

	TaskConvergenceStarted()
	TaskConvergenceDuration(duration time.Duration)
}

type taskStatMetronNotifier struct {
	clock                clock.Clock
	mutex                sync.Mutex
	metronClient         logging.IngressClient
	perCellStats         map[string]perCellStats
	globalTaskStats      globalStats
	convergenceRunsDelta uint64
	convergenceDuration  time.Duration
}

func NewTaskStatMetronNotifier(logger lager.Logger, clock clock.Clock, metronClient logging.IngressClient) TaskStatMetronNotifier {
	return &taskStatMetronNotifier{
		clock:        clock,
		perCellStats: make(map[string]perCellStats),
		metronClient: metronClient,
	}
}

func (t *taskStatMetronNotifier) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)
	for {
		select {
		case <-t.clock.NewTimer(60 * time.Second).C():
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
	t.metronClient.IncrementCounterWithDelta(PrunedTasksMetric, t.globalTaskStats.prunedTasks)
	t.metronClient.IncrementCounterWithDelta(KickedTasksMetric, t.globalTaskStats.kickedTasks)

	if t.convergenceRunsDelta > 0 {
		t.metronClient.IncrementCounterWithDelta(ConvergeTaskRunsCounter, t.convergenceRunsDelta)
		t.convergenceRunsDelta = 0
	}

	t.metronClient.SendDuration(ConvergeTaskDuration, t.convergenceDuration)
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

func (t *taskStatMetronNotifier) SnapshotTasks(pending, running, completed, resolving int, pruned, kicked uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.globalTaskStats.pendingTasks = pending
	t.globalTaskStats.runningTasks = running
	t.globalTaskStats.completedTasks = completed
	t.globalTaskStats.resolvingTasks = resolving
	t.globalTaskStats.prunedTasks = pruned
	t.globalTaskStats.kickedTasks = kicked
}

func (t *taskStatMetronNotifier) TaskConvergenceStarted() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.convergenceRunsDelta += 1
}

func (t *taskStatMetronNotifier) TaskConvergenceDuration(duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.convergenceDuration = duration
}
