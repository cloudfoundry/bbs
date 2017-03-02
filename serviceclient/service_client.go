package serviceclient

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	locketmodels "code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/rep/maintain"
	"golang.org/x/net/context"
)

const BBSLockSchemaKey = "bbs_lock"

//go:generate counterfeiter . ServiceClient

type ServiceClient interface {
	CellById(logger lager.Logger, cellId string) (*models.CellPresence, error)
	Cells(logger lager.Logger) (models.CellSet, error)
	CellEvents(logger lager.Logger) <-chan models.CellEvent
}

type serviceClient struct {
	cellPresenceClient maintain.CellPresenceClient
	locketClient       locketmodels.LocketClient
}

func NewServiceClient(cellPresenceClient maintain.CellPresenceClient, locketClient locketmodels.LocketClient) *serviceClient {
	return &serviceClient{
		cellPresenceClient: cellPresenceClient,
		locketClient:       locketClient,
	}
}

func (s *serviceClient) Cells(logger lager.Logger) (models.CellSet, error) {
	logger = logger.Session("cells")
	cellSet, err := s.cellPresenceClient.Cells(logger)
	if err != nil {
		logger.Error("failed-to-fetch-cells-from-consul", err)
		return nil, err
	}

	resp, err := s.locketClient.FetchAll(context.Background(), &locketmodels.FetchAllRequest{Type: locketmodels.PresenceType})
	if err != nil {
		logger.Error("failed-to-fetch-cells-from-locket", err)
		return nil, err
	}

	for _, resource := range resp.Resources {
		presence, err := presenceFromResource(resource)
		if err != nil {
			logger.Error("failed-to-unmarshal-presence", err)
			continue
		}
		cellSet.Add(presence)
	}

	return cellSet, nil
}

func (s *serviceClient) CellById(logger lager.Logger, cellId string) (*models.CellPresence, error) {
	logger = logger.Session("cell-by-id", lager.Data{"cell-id": cellId})

	presence, consulErr := s.cellPresenceClient.CellById(logger, cellId)
	if consulErr != nil {
		logger.Debug("failed-to-fetch-presence-from-consul", lager.Data{"error": consulErr})
	}

	resp, locketErr := s.locketClient.Fetch(context.Background(), &locketmodels.FetchRequest{
		Key: cellId,
	})
	if locketErr != nil {
		if consulErr != nil {
			logger.Error("failed-to-fetch-presence-from-locket", locketErr)
			if locketErr == locketmodels.ErrResourceNotFound {
				return nil, models.ErrResourceNotFound
			}
			return nil, locketErr
		}
		return presence, nil
	}

	var err error
	presence, err = presenceFromResource(resp.Resource)
	if err != nil {
		logger.Error("failed-to-unmarshal-presence", err)
		return nil, err
	}

	return presence, nil
}

func (s *serviceClient) CellEvents(logger lager.Logger) <-chan models.CellEvent {
	return s.cellPresenceClient.CellEvents(logger)
}

func presenceFromResource(resource *locketmodels.Resource) (*models.CellPresence, error) {
	cellPresence := &models.CellPresence{}
	err := json.Unmarshal([]byte(resource.Value), cellPresence)
	return cellPresence, err
}
