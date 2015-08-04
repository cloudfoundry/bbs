package auctionhandlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type LRPAuctionHandler struct {
	runner auctiontypes.AuctionRunner
}

func NewLRPAuctionHandler(runner auctiontypes.AuctionRunner) *LRPAuctionHandler {
	return &LRPAuctionHandler{
		runner: runner,
	}
}

func (*LRPAuctionHandler) logSession(logger lager.Logger) lager.Logger {
	return logger.Session("lrp-auction-handler")
}

func (h *LRPAuctionHandler) Create(w http.ResponseWriter, r *http.Request, logger lager.Logger) {
	logger = h.logSession(logger).Session("create")

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed-to-read-request-body", err)
		writeInternalErrorJSONResponse(w, err)
		return
	}

	starts := []models.LRPStartRequest{}
	err = json.Unmarshal(payload, &starts)
	if err != nil {
		logger.Error("malformed-json", err)
		writeInvalidJSONResponse(w, err)
		return
	}

	validStarts := make([]models.LRPStartRequest, 0, len(starts))
	lrpGuids := make(map[string][]uint)
	for _, start := range starts {
		if err := start.Validate(); err == nil {
			validStarts = append(validStarts, start)
			indices := lrpGuids[start.DesiredLRP.ProcessGuid]
			indices = append(indices, start.Indices...)
			lrpGuids[start.DesiredLRP.ProcessGuid] = indices
		} else {
			logger.Error("start-validate-failed", err, lager.Data{"lrp-start": start})
		}
	}

	h.runner.ScheduleLRPsForAuctions(validStarts)
	logger.Info("submitted", lager.Data{"lrps": lrpGuids})
	writeStatusAcceptedResponse(w)
}
