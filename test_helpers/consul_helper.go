package test_helpers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/gomega"
)

type ConsulHelper struct {
	consulClient consuladapter.Client
	logger       lager.Logger
}

func NewConsulHelper(logger lager.Logger, consulClient consuladapter.Client) *ConsulHelper {
	return &ConsulHelper{
		logger:       logger,
		consulClient: consulClient,
	}
}

func (t *ConsulHelper) RegisterCell(cell *models.CellPresence) {
	var err error
	jsonBytes, err := json.Marshal(cell)
	Expect(err).NotTo(HaveOccurred())

	// Use NewLock instead of NewPresence in order to block on the cell being registered
	runner := locket.NewLock(t.logger, t.consulClient, bbs.CellSchemaPath(cell.CellId), jsonBytes, clock.NewClock(), locket.RetryInterval, locket.LockTTL)
	ifrit.Invoke(runner)

	Expect(err).NotTo(HaveOccurred())
}
