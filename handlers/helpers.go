package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/gogo/protobuf/proto"
)

func writeUnknownErrorResponse(w http.ResponseWriter, err error) {
	writeProtoResponse(w, http.StatusInternalServerError, &bbs.Error{
		Type:    proto.String(bbs.UnknownError),
		Message: proto.String(err.Error()),
	})
}

func writeBadRequestResponse(w http.ResponseWriter, errorType string, err error) {
	writeJSONResponse(w, http.StatusBadRequest, bbs.Error{
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

func writeJSONResponse(w http.ResponseWriter, statusCode int, jsonObj interface{}) {
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		panic("Unable to encode JSON: " + err.Error())
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	w.Write(jsonBytes)
}

func writeEmptyResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(statusCode)
}
