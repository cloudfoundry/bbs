package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type ActualLRPHandler struct {
	db     db.ActualLRPDB
	logger lager.Logger
}

func NewActualLRPHandler(db db.ActualLRPDB, logger lager.Logger) *ActualLRPHandler {
	return &ActualLRPHandler{
		db:     db,
		logger: logger.Session("actuallrp-handler"),
	}
}

func (h *ActualLRPHandler) ActualLRPGroups(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	cellId := req.FormValue("cell_id")
	logger := h.logger.Session("actual-lrp-groups", lager.Data{
		"domain": domain, "cell_id": cellId,
	})

	filter := models.ActualLRPFilter{Domain: domain, CellID: cellId}
	actualLRPGroups, err := h.db.ActualLRPGroups(filter, h.logger)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-groups", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, actualLRPGroups)
}
