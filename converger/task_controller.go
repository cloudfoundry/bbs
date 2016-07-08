package converger

import "time"

//go:generate counterfeiter -o fake_handlers/fake_task_convergence_handler.go . TaskConvergenceHandler

type TaskController interface {
	ConvergeTasks(kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration) error
}
