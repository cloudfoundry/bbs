package test_helpers

import (
	"encoding/json"
	"fmt"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/onsi/gomega"
)

func (t *TestHelper) SetRawActualLRP(lrp models.ActualLRP) {
	value, err := json.Marshal(lrp) // do NOT use models.ToJSON; don't want validations
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.ActualLRPSchemaPath(lrp.GetProcessGuid(), lrp.GetIndex())
	_, err = t.etcdClient.Set(key, string(value), 0)

	Expect(err).NotTo(HaveOccurred())
}

func (t *TestHelper) SetRawEvacuatingActualLRP(lrp models.ActualLRP, ttlInSeconds uint64) {
	value, err := json.Marshal(lrp) // do NOT use models.ToJSON; don't want validations
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.EvacuatingActualLRPSchemaPath(lrp.GetProcessGuid(), lrp.GetIndex())
	_, err = t.etcdClient.Set(key, string(value), ttlInSeconds)

	Expect(err).NotTo(HaveOccurred())
}

func (t *TestHelper) CreateValidActualLRP(guid string, index int32) {
	t.SetRawActualLRP(t.NewValidActualLRP(guid, index))
}

func (t *TestHelper) CreateValidEvacuatingLRP(guid string, index int32) {
	t.SetRawEvacuatingActualLRP(t.NewValidActualLRP(guid, index), 100)
}

func (t *TestHelper) CreateMalformedActualLRP(guid string, index int32) {
	t.createMalformedValueForKey(etcddb.ActualLRPSchemaPath(guid, index))
}

func (t *TestHelper) CreateMalformedEvacuatingLRP(guid string, index int32) {
	t.createMalformedValueForKey(etcddb.EvacuatingActualLRPSchemaPath(guid, index))
}

func (t *TestHelper) createMalformedValueForKey(key string) {
	_, err := t.etcdClient.Create(key, "ßßßßßß", 0)

	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("error occurred at key '%s'", key))
}
