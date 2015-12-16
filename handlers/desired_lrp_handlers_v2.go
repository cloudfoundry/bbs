package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/models"
)

func (h *DesiredLRPHandler) DesiredLRPs_V2(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrps")

	request := &models.DesiredLRPsRequest{}
	response := &models.DesiredLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		filter := models.DesiredLRPFilter{Domain: request.Domain}
		response.DesiredLrps, err = h.db.DesiredLRPs(logger, filter)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesiredLRPByProcessGuid_V2(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrp-by-process-guid")

	request := &models.DesiredLRPByProcessGuidRequest{}
	response := &models.DesiredLRPResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		response.DesiredLrp, err = h.db.DesiredLRPByProcessGuid(logger, request.ProcessGuid)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
