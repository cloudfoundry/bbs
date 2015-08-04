package cellhandlers

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry-incubator/rep/lrp_stopper"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/pivotal-golang/lager"
)

type StopLRPInstanceHandler struct {
	logger  lager.Logger
	stopper lrp_stopper.LRPStopper
}

func NewStopLRPInstanceHandler(logger lager.Logger, stopper lrp_stopper.LRPStopper) *StopLRPInstanceHandler {
	return &StopLRPInstanceHandler{
		logger:  logger,
		stopper: stopper,
	}
}

func (h StopLRPInstanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	instanceGuid := r.FormValue(":instance_guid")

	logger := h.logger.Session("handling-stop-lrp-instance", lager.Data{
		"process-guid":  processGuid,
		"instance-guid": instanceGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if instanceGuid == "" {
		err := errors.New("instance_guid missing from request")
		logger.Error("missing-instance-guid", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go func() {
		err := h.stopper.StopInstance(processGuid, instanceGuid)
		if err == bbserrors.ErrStoreComparisonFailed {
			return
		}

		if err != nil {
			logger.Error("failed-to-stop", err)
			return
		}

		logger.Info("stopped")
	}()
}
