package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type CellHandler struct {
	logger        lager.Logger
	serviceClient bbs.ServiceClient
	exitChan      chan<- struct{}
}

func NewCellHandler(logger lager.Logger, serviceClient bbs.ServiceClient, exitChan chan<- struct{}) *CellHandler {
	return &CellHandler{
		logger:        logger.Session("cell-handler"),
		serviceClient: serviceClient,
		exitChan:      exitChan,
	}
}

func (h *CellHandler) Cells(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("cells")
	response := &models.CellsResponse{}
	cellSet, err := h.serviceClient.Cells(h.logger)
	cells := []*models.CellPresence{}
	for _, cp := range cellSet {
		cells = append(cells, cp)
	}
	response.Cells = cells
	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}
