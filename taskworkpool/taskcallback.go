package taskworkpool

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"regexp"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/pivotal-golang/lager"
)

const MAX_CB_RETRIES = 3
const TASK_CB_WORKERS = 20

//go:generate counterfeiter . TaskCompletionClient

type CompletedTaskHandler func(logger lager.Logger, taskDB db.TaskDB, task *models.Task)

type TaskCompletionClient interface {
	Submit(taskDB db.TaskDB, task *models.Task)
}

type TaskCompletionWorkPool struct {
	logger           lager.Logger
	callbackHandler  CompletedTaskHandler
	callbackWorkPool *workpool.WorkPool
}

func New(logger lager.Logger, cbHandler CompletedTaskHandler) *TaskCompletionWorkPool {
	if cbHandler == nil {
		panic("callbackHandler cannot be nil")
	}
	return &TaskCompletionWorkPool{
		logger:          logger,
		callbackHandler: cbHandler,
	}
}

func initializeWorkPool(logger lager.Logger) *workpool.WorkPool {
	cbWorkPool, err := workpool.NewWorkPool(TASK_CB_WORKERS)
	if err != nil {
		logger.Fatal("callback-workpool-creation-failed", err)
	}
	return cbWorkPool
}

func (twp *TaskCompletionWorkPool) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	twp.callbackWorkPool = initializeWorkPool(twp.logger)
	close(ready)

	<-signals
	go twp.callbackWorkPool.Stop()

	return nil
}

func (twp *TaskCompletionWorkPool) Submit(taskDB db.TaskDB, task *models.Task) {
	if twp.callbackWorkPool == nil {
		panic("called submit before workpool was started")
	}
	twp.callbackWorkPool.Submit(func() {
		twp.callbackHandler(twp.logger, taskDB, task)
	})
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
			CreatedAt:     task.CreatedAt,
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
				modelErr := taskDB.DeleteTask(logger, task.TaskGuid)
				if modelErr != nil {
					logger.Error("delete-task-failed", modelErr)
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
