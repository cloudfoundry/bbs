package handlers

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/locket/metrics/helpers"
)

type ActualLRPHandler struct {
	db             db.ActualLRPDB
	exitChan       chan<- struct{}
	requestMetrics helpers.RequestMetrics
	metricsGroup   string
}

func NewActualLRPHandler(db db.ActualLRPDB, exitChan chan<- struct{}, requestMetrics helpers.RequestMetrics) *ActualLRPHandler {
	return &ActualLRPHandler{
		db:             db,
		exitChan:       exitChan,
		requestMetrics: requestMetrics,
		metricsGroup:   "ActualLRPSEndpoint",
	}
}

func (h *ActualLRPHandler) ActualLRPs(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrps").WithTraceInfo(req)
	logger.Debug("starting")
	defer logger.Debug("complete")

	request := &models.ActualLRPsRequest{}
	response := &models.ActualLRPsResponse{}

	start := time.Now()
	startMetrics(h.requestMetrics, h.metricsGroup)
	defer stopMetrics(h.requestMetrics, h.metricsGroup, start, &err)

	err = parseRequest(logger, req, request)
	if err == nil {
		var index *int32
		if request.IndexExists() {
			i := request.GetIndex()
			index = &i
		}
		filter := models.ActualLRPFilter{Domain: request.Domain, CellID: request.CellId, Index: index, ProcessGuid: request.ProcessGuid}
		response.ActualLrps, err = h.db.ActualLRPs(req.Context(), logger, filter)
	}

	response.Error = models.ConvertError(err)

	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

// DEPRECATED
func (h *ActualLRPHandler) ActualLRPGroups(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups").WithTraceInfo(req)

	request := &models.ActualLRPGroupsRequest{}
	response := &models.ActualLRPGroupsResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
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

// DEPRECATED
func (h *ActualLRPHandler) ActualLRPGroupsByProcessGuid(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-groups-by-process-guid").WithTraceInfo(req)

	request := &models.ActualLRPGroupsByProcessGuidRequest{}
	response := &models.ActualLRPGroupsResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
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

// DEPRECATED
func (h *ActualLRPHandler) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("actual-lrp-group-by-process-guid-and-index").WithTraceInfo(req)

	request := &models.ActualLRPGroupByProcessGuidAndIndexRequest{}
	response := &models.ActualLRPGroupResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
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
