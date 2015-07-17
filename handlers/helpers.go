package handlers

import (
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/gogo/protobuf/proto"
)

func writeUnknownErrorResponse(w http.ResponseWriter, err error) {
	writeProtoResponse(w, http.StatusInternalServerError, &models.Error{
		Type:    proto.String(models.UnknownError),
		Message: proto.String(err.Error()),
	})
}

func writeNotFoundResponse(w http.ResponseWriter, err error) {
	writeProtoResponse(w, http.StatusNotFound, &models.Error{
		Type:    proto.String(models.ResourceNotFound),
		Message: proto.String(err.Error()),
	})
}

func writeBadRequestResponse(w http.ResponseWriter, errorType string, err error) {
	writeProtoResponse(w, http.StatusBadRequest, &models.Error{
		Type:    &errorType,
		Message: proto.String(err.Error()),
	})
}

func writeProtoResponse(w http.ResponseWriter, statusCode int, message proto.Message) {
	responseBytes, err := proto.Marshal(message)
	if err != nil {
		panic("Unable to encode Proto: " + err.Error())
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(responseBytes)))
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(statusCode)

	w.Write(responseBytes)
}

func writeEmptyResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(statusCode)
}
