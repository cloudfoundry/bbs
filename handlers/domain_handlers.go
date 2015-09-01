package handlers

import (
	"errors"
	"net/http"

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

func NewDomainHandler(logger lager.Logger, db db.DomainDB) *DomainHandler {
	return &DomainHandler{
		db:     db,
		logger: logger.Session("domain-handler"),
	}
}

func (h *DomainHandler) Domains(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("domains")
	response := &models.DomainsResponse{}
	response.Domains, err = h.db.Domains(logger)
	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}

func (h *DomainHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	var err error
	logger := h.logger.Session("upsert")

	request := &models.UpsertDomainRequest{}
	response := &models.UpsertDomainResponse{}

	err = parseRequest(logger, req, request)
	if err == nil {
		err = h.db.UpsertDomain(logger, request.Domain, request.Ttl)
	}

	response.Error = models.ConvertError(err)
	writeResponse(w, response)
}
