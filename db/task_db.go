package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type TaskFilter func(t *models.Task) bool

type CompleteTaskWork func(logger lager.Logger, taskDB TaskDB, task *models.Task) func()

//go:generate counterfeiter . TaskDB
type TaskDB interface {
	Tasks(logger lager.Logger, filter TaskFilter) (*models.Tasks, *models.Error)
	TaskByGuid(logger lager.Logger, processGuid string) (*models.Task, *models.Error)
	DesireTask(logger lager.Logger, request *models.DesireTaskRequest) *models.Error
	StartTask(logger lager.Logger, request *models.StartTaskRequest) (bool, *models.Error)
	CancelTask(logger lager.Logger, taskGuid string) *models.Error
	FailTask(logger lager.Logger, request *models.FailTaskRequest) *models.Error
	CompleteTask(logger lager.Logger, request *models.CompleteTaskRequest) *models.Error
	ResolvingTask(logger lager.Logger, taskGuid string) *models.Error
	ResolveTask(logger lager.Logger, taskGuid string) *models.Error
}
