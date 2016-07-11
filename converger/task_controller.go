package converger

import (
	"time"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o fake_controllers/fake_task_controller.go . TaskController

type TaskController interface {
	ConvergeTasks(logger lager.Logger, kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration) error
}
