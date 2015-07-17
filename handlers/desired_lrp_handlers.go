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

func NewDesiredLRPHandler(db db.DesiredLRPDB, logger lager.Logger) *DesiredLRPHandler {
	return &DesiredLRPHandler{
		db:     db,
		logger: logger.Session("desiredlrp-handler"),
	}
}

func (h *DesiredLRPHandler) DesiredLRPs(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("desired-lrps", lager.Data{
		"domain": domain,
	})

	desiredLRPs, err := h.db.DesiredLRPs(models.DesiredLRPFilter{Domain: domain}, h.logger)
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, desiredLRPs)
}

func (h *DesiredLRPHandler) DesiredLRPByProcessGuid(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("desired-lrps-process-guid", lager.Data{
		"process_guid": processGuid,
	})

	desiredLRP, err := h.db.DesiredLRPByProcessGuid(processGuid, h.logger)
	if err == models.ErrResourceNotFound {
		writeNotFoundResponse(w, err)
		return
	}
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrp", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, desiredLRP)
}
