package handlers

import (
	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"

	"golang.org/x/net/context"
)

//go:generate counterfeiter -o fake_controllers/fake_actual_lrp_lifecycle_controller.go . ActualLRPLifecycleController
type ActualLRPLifecycleController interface {
	ClaimActualLRP(logger lager.Logger, processGuid string, index int32, actualLRPInstanceKey *models.ActualLRPInstanceKey) error
	StartActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, actualLRPNetInfo *models.ActualLRPNetInfo) error
	CrashActualLRP(logger lager.Logger, actualLRPKey *models.ActualLRPKey, actualLRPInstanceKey *models.ActualLRPInstanceKey, errorMessage string) error
	FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error
	RemoveActualLRP(logger lager.Logger, processGuid string, index int32, instanceKey *models.ActualLRPInstanceKey) error
	RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error
}

type ActualLRPLifecycleHandler struct {
	db               db.ActualLRPDB
	desiredLRPDB     db.DesiredLRPDB
	actualHub        events.Hub
	auctioneerClient auctioneer.Client
	controller       ActualLRPLifecycleController
	exitChan         chan<- struct{}
}

func NewActualLRPLifecycleHandler(
	controller ActualLRPLifecycleController,
	exitChan chan<- struct{},
) *ActualLRPLifecycleHandler {
	return &ActualLRPLifecycleHandler{
		controller: controller,
		exitChan:   exitChan,
	}
}

func (h *bbsServer) ClaimActualLRP(context context.Context, req *models.ClaimActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("claim-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	err := h.actualLRPController.ClaimActualLRP(logger, req.ProcessGuid, req.Index, req.ActualLrpInstanceKey)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) StartActualLRP(context context.Context, request *models.StartActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("start-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	err := h.actualLRPController.StartActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) CrashActualLRP(context context.Context, request *models.CrashActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("crash-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	actualLRPKey := request.ActualLrpKey
	actualLRPInstanceKey := request.ActualLrpInstanceKey
	err := h.actualLRPController.CrashActualLRP(logger, actualLRPKey, actualLRPInstanceKey, request.ErrorMessage)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) FailActualLRP(ctx context.Context, request *models.FailActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("fail-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	err := h.actualLRPController.FailActualLRP(logger, request.ActualLrpKey, request.ErrorMessage)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) RemoveActualLRP(ctx context.Context, request *models.RemoveActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("remove-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	err := h.actualLRPController.RemoveActualLRP(logger, request.ProcessGuid, request.Index, request.ActualLrpInstanceKey)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) RetireActualLRP(ctx context.Context, request *models.RetireActualLRPRequest) (*models.ActualLRPLifecycleResponse, error) {
	logger := h.logger.Session("retire-actual-lrp")
	response := &models.ActualLRPLifecycleResponse{}
	err := h.actualLRPController.RetireActualLRP(logger, request.ActualLrpKey)
	response.Error = models.ConvertError(err)
	return response, nil
}
