package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

func (h *DesiredLRPHandler) DesiredLRPs_r0(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrps", lager.Data{"revision": 0})

	request := &models.DesiredLRPsRequest{}
	response := &models.DesiredLRPsResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		var lrps []*models.DesiredLRP

		filter := models.DesiredLRPFilter{Domain: request.Domain}
		lrps, err = h.db.DesiredLRPs(logger, filter)
		if err == nil {
			for i := range lrps {
				transformedLRP := lrps[i].VersionDownTo(format.V0)
				response.DesiredLrps = append(response.DesiredLrps, transformedLRP)
			}
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DesiredLRPHandler) DesiredLRPByProcessGuid_r0(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("desired-lrp-by-process-guid", lager.Data{"revision": 0})

	request := &models.DesiredLRPByProcessGuidRequest{}
	response := &models.DesiredLRPResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		var lrp *models.DesiredLRP
		lrp, err = h.db.DesiredLRPByProcessGuid(logger, request.ProcessGuid)
		if err == nil {
			transformedLRP := lrp.VersionDownTo(format.V0)
			response.DesiredLrp = transformedLRP
		}
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
