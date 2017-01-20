package handlers

import (
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"golang.org/x/net/context"
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

func (h *bbsServer) Cells(context context.Context, req *models.CellsRequest) (*models.CellsResponse, error) {
	logger := h.logger.Session("cells")
	response := &models.CellsResponse{}
	cellSet, err := h.serviceClient.Cells(logger)
	cells := []*models.CellPresence{}
	for _, cp := range cellSet {
		cells = append(cells, cp)
	}
	response.Cells = cells
	response.Error = models.ConvertError(err)
	return response, nil
}
