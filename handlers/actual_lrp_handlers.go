package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/v3"
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

func (h *ActualLRPHandler) ActualLRPs(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrps").WithTraceInfo(req)
	logger.Debug("starting")
	defer logger.Debug("complete")

	var request *models.ActualLRPsRequest
	protoRequest := &models.ProtoActualLRPsRequest{}
	response := &models.ActualLRPsResponse{}

	err = parseRequest(logger, req, protoRequest)
	request = protoRequest.FromProto()
	if err == nil {
		var index *int32
		if request.IndexExists() {
			i := request.GetIndex()
			index = i
		}
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId, Index: index, ProcessGuid: request.ProcessGuid}
		response.ActualLrps, err = h.db.ActualLRPs(req.Context(), logger, filter)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response.ToProto())
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

// Deprecated: use ActaulLRPs instead
func (h *ActualLRPHandler) ActualLRPGroups(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups").WithTraceInfo(req)

	var request *models.ActualLRPGroupsRequest
	protoRequest := &models.ProtoActualLRPGroupsRequest{}
	response := &models.ActualLRPGroupsResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer func() { writeResponse(w, response.ToProto()) }()

	err = parseRequest(logger, req, protoRequest)
	request = protoRequest.FromProto()
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId}
	lrps, err := h.db.ActualLRPs(req.Context(), logger, filter)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	response.ActualLrpGroups = models.ResolveActualLRPGroups(lrps)
}

// Deprecated: use ActaulLRPs instead
func (h *ActualLRPHandler) ActualLRPGroupsByProcessGuid(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups-by-process-guid").WithTraceInfo(req)

	var request *models.ActualLRPGroupsByProcessGuidRequest
	protoRequest := &models.ProtoActualLRPGroupsByProcessGuidRequest{}
	response := &models.ActualLRPGroupsResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer func() { writeResponse(w, response.ToProto()) }()

	err = parseRequest(logger, req, protoRequest)
	request = protoRequest.FromProto()
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	filter := models.ActualLRPFilter{ProcessGuid: request.ProcessGuid}
	lrps, err := h.db.ActualLRPs(req.Context(), logger, filter)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	response.ActualLrpGroups = models.ResolveActualLRPGroups(lrps)
}

// Deprecated: use ActaulLRPs instead
func (h *ActualLRPHandler) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-group-by-process-guid-and-index").WithTraceInfo(req)

	var request *models.ActualLRPGroupByProcessGuidAndIndexRequest
	protoRequest := &models.ProtoActualLRPGroupByProcessGuidAndIndexRequest{}
	response := &models.ActualLRPGroupResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer func() { writeResponse(w, response.ToProto()) }()

	err = parseRequest(logger, req, protoRequest)
	request = protoRequest.FromProto()
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	filter := models.ActualLRPFilter{ProcessGuid: request.ProcessGuid, Index: &request.Index}
	lrps, err := h.db.ActualLRPs(req.Context(), logger, filter)

	if err == nil && len(lrps) == 0 {
		err = models.ErrResourceNotFound
	}

	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}
	response.ActualLrpGroup = models.ResolveActualLRPGroup(lrps)
}
