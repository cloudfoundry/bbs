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
	logger := h.logger.Session("domains")
	response := &models.DomainsResponse{}
	response.Domains, response.Error = h.db.Domains(logger)
	writeResponse(w, response)
}

func (h *DomainHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("upsert")

	request := &models.UpsertDomainRequest{}
	response := &models.UpsertDomainResponse{}

	response.Error = parseRequest(logger, req, request)
	if response.Error == nil {
		response.Error = h.db.UpsertDomain(logger, request.Domain, request.Ttl)
	}

	writeResponse(w, response)
}
