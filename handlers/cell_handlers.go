package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
)

type CellHandler struct {
	serviceClient bbs.ServiceClient
	exitChan      chan<- struct{}
}

func NewCellHandler(serviceClient bbs.ServiceClient, exitChan chan<- struct{}) *CellHandler {
	return &CellHandler{
		serviceClient: serviceClient,
		exitChan:      exitChan,
	}
}

func (h *CellHandler) Cells(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("cells")
	response := &models.CellsResponse{}
	cellSet, err := h.serviceClient.Cells(logger)
	cells := []*models.CellPresence{}
	for _, cp := range cellSet {
		cells = append(cells, cp)
	}
	response.Cells = cells
	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}

func (h *CellHandler) Cells_r1(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("cells")
	response := &models.CellsResponse{}
	cellSet, err := h.serviceClient.Cells(logger)
	cells := []*models.CellPresence{}
	for _, cp := range cellSet {
		cells = append(cells, cp.VersionDownTo(format.V1))
	}
	response.Cells = cells
	response.Error = models.ConvertError(err)
	writeResponse(w, response)
	exitIfUnrecoverable(logger, h.exitChan, response.Error)
}
