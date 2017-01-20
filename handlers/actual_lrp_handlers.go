package handlers

import (
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"golang.org/x/net/context"
)

type ActualLRPHandler struct {
	db       db.ActualLRPDB
	exitChan chan<- struct{}
}

func NewActualLRPHandler(db db.ActualLRPDB, exitChan chan<- struct{}) *ActualLRPHandler {
	return &ActualLRPHandler{
		db:       db,
		exitChan: exitChan,
	}
}

func (h *bbsServer) ActualLRPGroups(
	context context.Context,
	req *models.ActualLRPGroupsRequest,
) (*models.ActualLRPGroupsResponse, error) {
	var err error
	logger := h.logger.Session("actual-lrp-groups")

	response := &models.ActualLRPGroupsResponse{}

	filter := models.ActualLRPFilter{Domain: req.Domain, CellID: req.CellId}
	response.ActualLrpGroups, err = h.db.ActualLRPGroups(logger, filter)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) ActualLRPGroupsByProcessGuid(
	context context.Context,
	req *models.ActualLRPGroupsByProcessGuidRequest,
) (*models.ActualLRPGroupsResponse, error) {
	var err error
	logger := h.logger.Session("actual-lrp-groups-by-process-guid")

	response := &models.ActualLRPGroupsResponse{}

	response.ActualLrpGroups, err = h.db.ActualLRPGroupsByProcessGuid(logger, req.ProcessGuid)
	response.Error = models.ConvertError(err)
	return response, nil
}

func (h *bbsServer) ActualLRPGroupByProcessGuidAndIndex(
	context context.Context,
	req *models.ActualLRPGroupByProcessGuidAndIndexRequest,
) (*models.ActualLRPGroupResponse, error) {
	logger := h.logger.Session("actual-lrp-group-by-process-guid-and-index")

	actualLRPGroup, err := h.db.ActualLRPGroupByProcessGuidAndIndex(logger, req.ProcessGuid, req.Index)

	response := &models.ActualLRPGroupResponse{}
	response.Error = models.ConvertError(err)
	response.ActualLrpGroup = actualLRPGroup

	return response, nil
}
