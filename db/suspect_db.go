package db

import (
	"context"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//counterfeiter:generate . SuspectDB

type SuspectDB interface {
	RemoveSuspectActualLRP(context.Context, lager.Logger, *models.ActualLRPKey) (*models.ActualLRP, error)
}
