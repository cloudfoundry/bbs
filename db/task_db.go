package db

import (
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type TaskFilter func(t *models.Task) bool

type CompleteTaskWork func(logger lager.Logger, taskDB TaskDB, task *models.Task) func()

//go:generate counterfeiter . TaskDB
type TaskDB interface {
	Tasks(logger lager.Logger, filter TaskFilter) (*models.Tasks, *models.Error)
	TaskByGuid(logger lager.Logger, processGuid string) (*models.Task, *models.Error)

	DesireTask(logger lager.Logger, taskDefinition *models.TaskDefinition, taskGuid, domain string) *models.Error
	StartTask(logger lager.Logger, taskGuid, cellId string) (bool, *models.Error)
	CancelTask(logger lager.Logger, taskGuid string) *models.Error
	FailTask(logger lager.Logger, taskGuid, failureReason string) *models.Error
	CompleteTask(logger lager.Logger, taskGuid, cellId string, failed bool, failureReason, result string) *models.Error
	ResolvingTask(logger lager.Logger, taskGuid string) *models.Error
	ResolveTask(logger lager.Logger, taskGuid string) *models.Error

	ConvergeTasks(
		logger lager.Logger,
		kickTaskDuration, expirePendingTaskDuration, expireCompletedTaskDuration time.Duration,
	)
}
