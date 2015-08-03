package db

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type TaskFilter func(t *models.Task) bool

//go:generate counterfeiter . TaskDB
type TaskDB interface {
	Tasks(logger lager.Logger, filter TaskFilter) (*models.Tasks, *models.Error)
	TaskByGuid(logger lager.Logger, processGuid string) (*models.Task, *models.Error)
	DesireTask(logger lager.Logger, guid, domain string, def *models.TaskDefinition) error
}
