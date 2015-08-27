package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type LRPConvergenceHandler struct {
	logger lager.Logger
	db     db.LRPDB
}

func NewLRPConvergenceHandler(logger lager.Logger, db db.LRPDB) *LRPConvergenceHandler {
	return &LRPConvergenceHandler{logger, db}
}

func (h *LRPConvergenceHandler) ConvergeLRPs(w http.ResponseWriter, req *http.Request) {
	h.db.ConvergeLRPs(h.logger)

	response := &models.ConvergeLRPsResponse{}
	writeResponse(w, response)
}
