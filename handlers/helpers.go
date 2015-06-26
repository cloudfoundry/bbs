package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
)

func writeUnknownErrorResponse(w http.ResponseWriter, err error) {
	writeJSONResponse(w, http.StatusInternalServerError, bbs.Error{
		Type:    bbs.UnknownError,
		Message: err.Error(),
	})
}

func writeBadRequestResponse(w http.ResponseWriter, errorType string, err error) {
	writeJSONResponse(w, http.StatusBadRequest, bbs.Error{
		Type:    errorType,
		Message: err.Error(),
	})
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
