package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type CellHandler struct {
	logger        lager.Logger
	serviceClient bbs.ServiceClient
}

func NewCellHandler(logger lager.Logger, serviceClient bbs.ServiceClient) *CellHandler {
	return &CellHandler{
		logger:        logger.Session("cell-handler"),
		serviceClient: serviceClient,
	}
}

func (h *CellHandler) Cells(w http.ResponseWriter, req *http.Request) {
	var err error
	h.logger.Session("cells")
	response := &models.CellsResponse{}
	cellSet, err := h.serviceClient.Cells(h.logger)
	cells := []*models.CellPresence{}
	for _, cp := range cellSet {
		cells = append(cells, cp)
	}
	response.Cells = cells
	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
