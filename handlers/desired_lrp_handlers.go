package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
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
	logger := h.logger.Session("desired-lrp-groups", lager.Data{})

	desiredLRPs, err := h.db.DesiredLRPs(h.logger)
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, desiredLRPs)
}
