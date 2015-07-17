package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/pivotal-golang/lager"
)

type DomainHandler struct {
	db     db.DomainDB
	logger lager.Logger
}

var (
	ErrDomainMissing = errors.New("domain missing from request")
	ErrMaxAgeMissing = errors.New("max-age directive missing from request")
)

func NewDomainHandler(db db.DomainDB, logger lager.Logger) *DomainHandler {
	return &DomainHandler{
		db:     db,
		logger: logger.Session("domain-handler"),
	}
}

func (h *DomainHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	domains, err := h.db.GetAllDomains(logger)
	if err != nil {
		writeUnknownErrorResponse(w, err)
		return
	}

	writeProtoResponse(w, http.StatusOK, domains)
}

func (h *DomainHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("upsert")
	domain := req.FormValue(":domain")

	ttl := 0

	cacheControl := req.Header["Cache-Control"]
	if cacheControl != nil {
		var maxAge string
		for _, directive := range cacheControl {
			if strings.HasPrefix(directive, "max-age=") {
				maxAge = directive
				break
			}
		}
		if maxAge == "" {
			logger.Error("missing-max-age-directive", ErrMaxAgeMissing)
			writeBadRequestResponse(w, models.InvalidRequest, ErrMaxAgeMissing)
			return
		}

		var err error
		ttl, err = strconv.Atoi(maxAge[8:])
		if err != nil {
			err := fmt.Errorf("invalid-max-age-directive: %s", maxAge)
			logger.Error("invalid-max-age-directive", err)
			writeBadRequestResponse(w, models.InvalidRequest, err)
			return
		}
	}

	err := h.db.UpsertDomain(domain, ttl, logger)
	if err != nil {
		writeUnknownErrorResponse(w, err)
		return
	}

	writeEmptyResponse(w, http.StatusNoContent)
}
