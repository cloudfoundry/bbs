package etcd

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

const throttlerSize = 20

const (
	convergeTaskRunsCounter = metric.Counter("ConvergenceTaskRuns")
	convergeTaskDuration    = metric.Duration("ConvergenceTaskDuration")

	tasksKickedCounter = metric.Counter("ConvergenceTasksKicked")
	tasksPrunedCounter = metric.Counter("ConvergenceTasksPruned")
)

type compareAndSwappableTask struct {
	OldIndex uint64
	NewTask  *models.Task
}

func (db *ETCDDB) ConvergeTasks(
	logger lager.Logger,
	kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration,
) {
	logger = logger.Session("converge-tasks")
	logger.Info("starting-convergence")
	defer logger.Info("finished-convergence")

	convergeTaskRunsCounter.Increment()

	convergeStart := db.clock.Now()

	defer func() {
		convergeTaskDuration.Send(time.Since(convergeStart))
	}()

	logger.Debug("listing-tasks")
	taskState, modelErr := db.fetchRecursiveRaw(logger, TaskSchemaRoot)
	if modelErr != nil {
		logger.Debug("failed-listing-task")
		return
	}
	logger.Debug("succeeded-listing-task")

	cellsLoader := db.cellDB.NewCellsLoader(logger)
	logger.Debug("listing-cells")
	cellSet, modelErr := cellsLoader.Cells()
	if modelErr != nil {
		if !models.ErrResourceNotFound.Equal(modelErr) {
			logger.Debug("failed-listing-cells")
			return
		}

		cellSet = models.CellSet{}
	}

	logger.Debug("succeeded-listing-cells")

	logError := func(task *models.Task, message string) {
		logger.Error(message, nil, lager.Data{
			"task-guid": task.TaskGuid,
		})
	}

	tasksToComplete := []*models.Task{}
	scheduleForCompletion := func(task *models.Task) {
		if task.CompletionCallbackUrl == "" {
			return
		}
		tasksToComplete = append(tasksToComplete, task)
	}

	keysToDelete := []string{}

	tasksToCAS := []compareAndSwappableTask{}
	scheduleForCASByIndex := func(index uint64, newTask *models.Task) {
		tasksToCAS = append(tasksToCAS, compareAndSwappableTask{
			OldIndex: index,
			NewTask:  newTask,
		})
	}

	tasksToAuction := []*models.Task{}

	var tasksKicked uint64 = 0

	logger.Debug("determining-convergence-work", lager.Data{"num-tasks": len(taskState.Nodes)})
	for _, node := range taskState.Nodes {
		task := new(models.Task)
		err := db.deserializeModel(logger, node, task)
		if err != nil {
			logger.Error("failed-to-unmarshal-task-json", err, lager.Data{
				"key":   node.Key,
				"value": node.Value,
			})

			keysToDelete = append(keysToDelete, node.Key)
			continue
		}

		shouldKickTask := db.durationSinceTaskUpdated(task) >= kickTaskDuration

		switch task.State {
		case models.Task_Pending:
			shouldMarkAsFailed := db.durationSinceTaskCreated(task) >= expirePendingTaskDuration
			if shouldMarkAsFailed {
				logError(task, "failed-to-start-in-time")
				db.markTaskFailed(task, "not started within time limit")
				scheduleForCASByIndex(node.ModifiedIndex, task)
				tasksKicked++
			} else if shouldKickTask {
				logger.Info("requesting-auction-for-pending-task", lager.Data{"task-guid": task.TaskGuid})
				tasksToAuction = append(tasksToAuction, task)
				tasksKicked++
			}
		case models.Task_Running:
			cellIsAlive := cellSet.HasCellID(task.CellId)
			if !cellIsAlive {
				logError(task, "cell-disappeared")
				db.markTaskFailed(task, "cell disappeared before completion")
				scheduleForCASByIndex(node.ModifiedIndex, task)
				tasksKicked++
			}
		case models.Task_Completed:
			shouldDeleteTask := db.durationSinceTaskFirstCompleted(task) >= expireCompletedTaskDuration
			if shouldDeleteTask {
				logError(task, "failed-to-start-resolving-in-time")
				keysToDelete = append(keysToDelete, node.Key)
			} else if shouldKickTask {
				logger.Info("kicking-completed-task", lager.Data{"task-guid": task.TaskGuid})
				scheduleForCompletion(task)
				tasksKicked++
			}
		case models.Task_Resolving:
			shouldDeleteTask := db.durationSinceTaskFirstCompleted(task) >= expireCompletedTaskDuration
			if shouldDeleteTask {
				logError(task, "failed-to-resolve-in-time")
				keysToDelete = append(keysToDelete, node.Key)
			} else if shouldKickTask {
				logger.Info("demoting-resolving-to-completed", lager.Data{"task-guid": task.TaskGuid})
				demoted := demoteToCompleted(task)
				scheduleForCASByIndex(node.ModifiedIndex, demoted)
				scheduleForCompletion(demoted)
				tasksKicked++
			}
		}
	}
	logger.Debug("done-determining-convergence-work", lager.Data{
		"num-tasks-to-auction":  len(tasksToAuction),
		"num-tasks-to-cas":      len(tasksToCAS),
		"num-tasks-to-complete": len(tasksToComplete),
		"num-keys-to-delete":    len(keysToDelete),
	})

	if len(tasksToAuction) > 0 {
		logger.Debug("requesting-task-auctions", lager.Data{"num-tasks-to-auction": len(tasksToAuction)})
		if err := db.auctioneerClient.RequestTaskAuctions(tasksToAuction); err != nil {
			taskGuids := make([]string, len(tasksToAuction))
			for i, task := range tasksToAuction {
				taskGuids[i] = task.TaskGuid
			}
			logger.Error("failed-to-request-auctions-for-pending-tasks", err,
				lager.Data{"task-guids": taskGuids})
		}
		logger.Debug("done-requesting-task-auctions", lager.Data{"num-tasks-to-auction": len(tasksToAuction)})
	}

	tasksKickedCounter.Add(tasksKicked)
	logger.Debug("compare-and-swapping-tasks", lager.Data{"num-tasks-to-cas": len(tasksToCAS)})
	err := db.batchCompareAndSwapTasks(tasksToCAS, logger)
	if err != nil {
		return
	}
	logger.Debug("done-compare-and-swapping-tasks", lager.Data{"num-tasks-to-cas": len(tasksToCAS)})

	logger.Debug("submitting-tasks-to-be-completed", lager.Data{"num-tasks-to-complete": len(tasksToComplete)})
	for _, task := range tasksToComplete {
		db.taskCompletionClient.Submit(db, task)
	}
	logger.Debug("done-submitting-tasks-to-be-completed", lager.Data{"num-tasks-to-complete": len(tasksToComplete)})

	tasksPrunedCounter.Add(uint64(len(keysToDelete)))
	logger.Debug("deleting-keys", lager.Data{"num-keys-to-delete": len(keysToDelete)})
	db.batchDeleteTasks(keysToDelete, logger)
	logger.Debug("done-deleting-keys", lager.Data{"num-keys-to-delete": len(keysToDelete)})
}

func (db *ETCDDB) durationSinceTaskCreated(task *models.Task) time.Duration {
	return db.clock.Now().Sub(time.Unix(0, task.CreatedAt))
}

func (db *ETCDDB) durationSinceTaskUpdated(task *models.Task) time.Duration {
	return db.clock.Now().Sub(time.Unix(0, task.UpdatedAt))
}

func (db *ETCDDB) durationSinceTaskFirstCompleted(task *models.Task) time.Duration {
	if task.FirstCompletedAt == 0 {
		return 0
	}
	return db.clock.Now().Sub(time.Unix(0, task.FirstCompletedAt))
}

func (db *ETCDDB) markTaskFailed(task *models.Task, reason string) {
	db.markTaskCompleted(task, true, reason, "")
}

func demoteToCompleted(task *models.Task) *models.Task {
	task.State = models.Task_Completed
	return task
}

func (db *ETCDDB) batchCompareAndSwapTasks(tasksToCAS []compareAndSwappableTask, logger lager.Logger) error {
	if len(tasksToCAS) == 0 {
		return nil
	}

	works := []func(){}

	for _, taskToCAS := range tasksToCAS {
		task := taskToCAS.NewTask
		task.UpdatedAt = db.clock.Now().UnixNano()
		value, err := db.serializeModel(logger, task)
		if err != nil {
			logger.Error("failed-to-marshal", err, lager.Data{
				"task-guid": task.TaskGuid,
			})
			continue
		}

		index := taskToCAS.OldIndex
		works = append(works, func() {
			_, err := db.client.CompareAndSwap(TaskSchemaPathByGuid(task.TaskGuid), value, NO_TTL, index)
			if err != nil {
				logger.Error("failed-to-compare-and-swap", err, lager.Data{
					"task-guid": task.TaskGuid,
				})
			}
		})
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		return err
	}

	throttler.Work()
	return nil
}

func (db *ETCDDB) batchDeleteTasks(taskGuids []string, logger lager.Logger) {
	if len(taskGuids) == 0 {
		return
	}

	works := []func(){}

	for _, taskGuid := range taskGuids {
		taskGuid := taskGuid
		works = append(works, func() {
			_, err := db.client.Delete(taskGuid, true)
			if err != nil {
				logger.Error("failed-to-delete", err, lager.Data{
					"task-guid": taskGuid,
				})
			}
		})
	}

	throttler, err := workpool.NewThrottler(throttlerSize, works)
	if err != nil {
		logger.Error("failed-to-create-throttler", err)
	}

	throttler.Work()
	return
}
