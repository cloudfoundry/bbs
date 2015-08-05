package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type EvacuationHandler struct {
	db     db.EvacuationDB
	logger lager.Logger
}

func NewEvacuationHandler(logger lager.Logger, db db.EvacuationDB) *EvacuationHandler {
	return &EvacuationHandler{
		db:     db,
		logger: logger.Session("evacuation-handler"),
	}
}

func (h *EvacuationHandler) RemoveEvacuatingActualLRP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("remove-evacuating-actual-lrp")

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	request := &models.RemoveEvacuatingActualLRPRequest{}
	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}
	logger.Debug("parsed-request-body", lager.Data{"request": request})
	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		writeBadRequestResponse(w, models.InvalidRequest, err)
		return
	}

	bbsErr := h.db.RemoveEvacuatingActualLRP(logger, request)
	if bbsErr != nil {
		logger.Error("failed-to-retire-actual-lrp", bbsErr)
		if bbsErr.Equal(models.ErrResourceNotFound) {
			writeNotFoundResponse(w, bbsErr)
		} else {
			writeUnknownErrorResponse(w, bbsErr)
		}
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}
