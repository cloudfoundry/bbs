package etcd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/pivotal-golang/lager"
)

const MAX_CB_RETRIES = 3

func CompleteTaskWork(logger lager.Logger, taskDB db.TaskDB, task *models.Task) func() {
	return func() { HandleCompletedTask(logger, taskDB, task) }
}

func HandleCompletedTask(logger lager.Logger, taskDB db.TaskDB, task *models.Task) {
	logger = logger.WithData(lager.Data{"task-guid": task.TaskGuid})

	httpClient := cf_http.NewClient()

	if task.CompletionCallbackUrl != "" {
		logger.Info("resolving-task")
		modelErr := taskDB.ResolvingTask(logger, task.TaskGuid)
		if modelErr != nil {
			logger.Error("marking-task-as-resolving-failed", modelErr)
			return
		}

		logger = logger.WithData(lager.Data{"callback_url": task.CompletionCallbackUrl})

		json, err := json.Marshal(&models.TaskCallbackResponse{
			TaskGuid:      task.TaskGuid,
			Failed:        task.Failed,
			FailureReason: task.FailureReason,
			Result:        task.Result,
			Annotation:    task.Annotation,
		})
		if err != nil {
			logger.Error("marshalling-task-failed", err)
			return
		}

		var statusCode int

		for i := 0; i < MAX_CB_RETRIES; i++ {
			request, err := http.NewRequest("POST", task.CompletionCallbackUrl, bytes.NewReader(json))
			if err != nil {
				logger.Error("building-request-failed", err)
				return
			}

			request.Header.Set("Content-Type", "application/json")
			response, err := httpClient.Do(request)
			if err != nil {
				matched, _ := regexp.MatchString("use of closed network connection", err.Error())
				if matched {
					continue
				}
				logger.Error("doing-request-failed", err)
				return
			}
			defer response.Body.Close()

			statusCode = response.StatusCode
			if shouldResolve(statusCode) {
				modelErr := taskDB.ResolveTask(logger, task.TaskGuid)
				if modelErr != nil {
					logger.Error("resolve-task-failed", modelErr)
					return
				}

				logger.Info("resolved-task", lager.Data{"status_code": statusCode})
				return
			}
		}

		logger.Info("callback-failed", lager.Data{"status_code": statusCode})
	}
	return
}

func shouldResolve(status int) bool {
	switch status {
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return false
	default:
		return true
	}
}
