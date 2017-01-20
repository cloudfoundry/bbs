package handlers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/db"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/taskworkpool"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/rep"
	"github.com/gogo/protobuf/proto"
)

type bbsServer struct {
	logger              lager.Logger
	db                  db.DB
	taskController      TaskController
	actualLRPController ActualLRPLifecycleController
	serviceClient       bbs.ServiceClient
}

func New(
	logger, accessLogger lager.Logger,
	updateWorkers int,
	convergenceWorkersSize int,
	db db.DB,
	desiredHub, actualHub events.Hub,
	taskCompletionClient taskworkpool.TaskCompletionClient,
	serviceClient bbs.ServiceClient,
	auctioneerClient auctioneer.Client,
	repClientFactory rep.ClientFactory,
	migrationsDone <-chan struct{},
	exitChan chan struct{},
) *bbsServer {
	// pingHandler := NewPingHandler()
	// domainHandler := NewDomainHandler(db, exitChan)
	// actualLRPHandler := NewActualLRPHandler(db, exitChan)
	actualLRPController := controllers.NewActualLRPLifecycleController(db, db, db, auctioneerClient, serviceClient, repClientFactory, actualHub)
	// actualLRPLifecycleHandler := NewActualLRPLifecycleHandler(actualLRPController, exitChan)
	// evacuationHandler := NewEvacuationHandler(db, db, db, actualHub, auctioneerClient, exitChan)
	// desiredLRPHandler := NewDesiredLRPHandler(updateWorkers, db, db, desiredHub, actualHub, auctioneerClient, repClientFactory, serviceClient, exitChan)
	taskController := controllers.NewTaskController(db, taskCompletionClient, auctioneerClient, serviceClient, repClientFactory)
	// taskHandler := NewTaskHandler(taskController, exitChan)
	// eventsHandler := NewEventHandler(desiredHub, actualHub)
	// cellsHandler := NewCellHandler(serviceClient, exitChan)

	return &bbsServer{
		logger:              logger,
		taskController:      taskController,
		actualLRPController: actualLRPController,
		serviceClient:       serviceClient,
	}
}

func route(f http.HandlerFunc) http.Handler {
	return f
}

func parseRequest(logger lager.Logger, req *http.Request, request MessageValidator) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		return models.ErrUnknownError
	}

	err = request.Unmarshal(data)
	if err != nil {
		logger.Error("failed-to-parse-request-body", err)
		return models.ErrBadRequest
	}

	if err := request.Validate(); err != nil {
		logger.Error("invalid-request", err)
		return models.NewError(models.Error_InvalidRequest, err.Error())
	}

	return nil
}

func exitIfUnrecoverable(logger lager.Logger, exitCh chan<- struct{}, err *models.Error) {
	if err != nil && err.Type == models.Error_Unrecoverable {
		logger.Error("unrecoverable-error", err)
		select {
		case exitCh <- struct{}{}:
		default:
		}
	}
}

func writeResponse(w http.ResponseWriter, message proto.Message) {
	responseBytes, err := proto.Marshal(message)
	if err != nil {
		panic("Unable to encode Proto: " + err.Error())
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(responseBytes)))
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)

	w.Write(responseBytes)
}
