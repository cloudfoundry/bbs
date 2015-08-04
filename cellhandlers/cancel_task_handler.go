package cellhandlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/executor"
	"github.com/pivotal-golang/lager"
)

type CancelTaskHandler struct {
	logger         lager.Logger
	executorClient executor.Client
}

func NewCancelTaskHandler(logger lager.Logger, executorClient executor.Client) *CancelTaskHandler {
	return &CancelTaskHandler{
		logger:         logger,
		executorClient: executorClient,
	}
}

func (h CancelTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	taskGuid := r.FormValue(":task_guid")

	logger := h.logger.Session("cancel-task", lager.Data{
		"instance-guid": taskGuid,
	})

	w.WriteHeader(http.StatusAccepted)

	go func() {
		logger.Info("deleting-container")
		err := h.executorClient.DeleteContainer(taskGuid)
		if err == executor.ErrContainerNotFound {
			logger.Info("container-not-found")
			return
		}

		if err != nil {
			logger.Error("failed-deleting-container", err)
			return
		}

		logger.Info("succeeded-deleting-container")
	}()
}
