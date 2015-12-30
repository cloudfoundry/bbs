package test_helpers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"

	. "github.com/onsi/gomega"
)

type ConsulHelper struct {
	consulSession *consuladapter.Session
}

func NewConsulHelper(consulSession *consuladapter.Session) *ConsulHelper {
	return &ConsulHelper{consulSession: consulSession}
}

func (t *ConsulHelper) RegisterCell(cell *models.CellPresence) {
	var err error
	jsonBytes, err := json.Marshal(cell)
	Expect(err).NotTo(HaveOccurred())

	err = t.consulSession.AcquireLock(bbs.CellSchemaPath(cell.CellId), jsonBytes)
	Expect(err).NotTo(HaveOccurred())
}
