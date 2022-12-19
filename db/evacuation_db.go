package db

import (
	"context"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

//counterfeiter:generate . EvacuationDB

type EvacuationDB interface {
	RemoveEvacuatingActualLRP(context.Context, lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey) error
	EvacuateActualLRP(context.Context, lager.Logger, *models.ActualLRPKey, *models.ActualLRPInstanceKey, *models.ActualLRPNetInfo, []*models.ActualLRPInternalRoute) (actualLRP *models.ActualLRP, err error)
}
