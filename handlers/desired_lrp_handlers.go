package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type DesiredLRPHandler struct {
	db     db.DesiredLRPDB
	logger lager.Logger
}

func NewDesiredLRPHandler(logger lager.Logger, db db.DesiredLRPDB) *DesiredLRPHandler {
	return &DesiredLRPHandler{
		db:     db,
		logger: logger.Session("desiredlrp-handler"),
	}
}

func (h *DesiredLRPHandler) DesiredLRPs(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrps")

	request := &models.DesiredLRPsRequest{}
	response := &models.DesiredLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		var lrps []*models.DesiredLRP

		filter := models.DesiredLRPFilter{Domain: request.Domain}
		lrps, err = h.db.DesiredLRPs(logger, filter)
		if err == nil {
			for i := range lrps {
				transformedLRP := lrps[i].WithCacheDependenciesAsSetupActions()
				response.DesiredLrps = append(response.DesiredLrps, &transformedLRP)
			}
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesiredLRPByProcessGuid(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrp-by-process-guid")

	request := &models.DesiredLRPByProcessGuidRequest{}
	response := &models.DesiredLRPResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		var lrp *models.DesiredLRP
		lrp, err = h.db.DesiredLRPByProcessGuid(logger, request.ProcessGuid)
		if err == nil {
			transformedLRP := lrp.WithCacheDependenciesAsSetupActions()
			response.DesiredLrp = &transformedLRP
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesiredLRPSchedulingInfos(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrps")

	request := &models.DesiredLRPsRequest{}
	response := &models.DesiredLRPSchedulingInfosResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.DesiredLRPFilter{Domain: request.Domain}
		response.DesiredLrpSchedulingInfos, err = h.db.DesiredLRPSchedulingInfos(logger, filter)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesireDesiredLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desire-lrp")

	request := &models.DesireLRPRequest{}
	response := &models.DesiredLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.DesireLRP(logger, request.DesiredLrp)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) UpdateDesiredLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("update-desired-lrp")

	request := &models.UpdateDesiredLRPRequest{}
	response := &models.DesiredLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.UpdateDesiredLRP(logger, request.ProcessGuid, request.Update)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) RemoveDesiredLRP(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("remove-desired-lrp")

	request := &models.RemoveDesiredLRPRequest{}
	response := &models.DesiredLRPLifecycleResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.RemoveDesiredLRP(logger, request.ProcessGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
