package consul_helpers

import (
	"github.com/cloudfoundry-incubator/bbs/db/consul"
	"github.com/cloudfoundry-incubator/bbs/models"

	. "github.com/onsi/gomega"
)

func (t *ConsulHelper) RegisterCell(cell models.CellPresence) {
	var err error
	jsonBytes, err := models.ToJSON(cell)
	Expect(err).NotTo(HaveOccurred())

	err = t.consulSession.AcquireLock(consul.CellSchemaPath(cell.CellID), jsonBytes)
	Expect(err).NotTo(HaveOccurred())
}
