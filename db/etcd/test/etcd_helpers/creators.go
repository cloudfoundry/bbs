package etcd_helpers

import (
	"fmt"

	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/models/test/model_helpers"
	. "github.com/onsi/gomega"
)

func (t *ETCDHelper) SetRawActualLRP(lrp *models.ActualLRP) {
	value, err := t.serializer.Marshal(t.logger, t.format, lrp)
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.ActualLRPSchemaPath(lrp.GetProcessGuid(), lrp.GetIndex())
	_, err = t.client.Set(key, value, 0)

	Expect(err).NotTo(HaveOccurred())
}

func (t *ETCDHelper) SetRawEvacuatingActualLRP(lrp *models.ActualLRP, ttlInSeconds uint64) {
	value, err := t.serializer.Marshal(t.logger, t.format, lrp)
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.EvacuatingActualLRPSchemaPath(lrp.GetProcessGuid(), lrp.GetIndex())
	_, err = t.client.Set(key, value, ttlInSeconds)

	Expect(err).NotTo(HaveOccurred())
}

func (t *ETCDHelper) SetRawDesiredLRP(lrp *models.DesiredLRP) {
	value, err := t.serializer.Marshal(t.logger, t.format, lrp)
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.DesiredLRPSchemaPath(lrp)
	_, err = t.client.Set(key, value, 0)

	Expect(err).NotTo(HaveOccurred())
}

func (t *ETCDHelper) SetRawTask(task *models.Task) {
	value, err := t.serializer.Marshal(t.logger, t.format, task)
	Expect(err).NotTo(HaveOccurred())

	key := etcddb.TaskSchemaPath(task)
	_, err = t.client.Set(key, value, 0)

	Expect(err).NotTo(HaveOccurred())
}

func (t *ETCDHelper) SetRawDomain(domain string) {
	_, err := t.client.Set(etcddb.DomainSchemaPath(domain), []byte{}, 0)
	Expect(err).NotTo(HaveOccurred())
}

func (t *ETCDHelper) CreateValidActualLRP(guid string, index int32) {
	t.SetRawActualLRP(model_helpers.NewValidActualLRP(guid, index))
}

func (t *ETCDHelper) CreateValidEvacuatingLRP(guid string, index int32) {
	t.SetRawEvacuatingActualLRP(model_helpers.NewValidActualLRP(guid, index), 100)
}

func (t *ETCDHelper) CreateValidDesiredLRP(guid string) {
	t.SetRawDesiredLRP(model_helpers.NewValidDesiredLRP(guid))
}

func (t *ETCDHelper) CreateValidTask(guid string) {
	t.SetRawTask(model_helpers.NewValidTask(guid))
}

func (t *ETCDHelper) CreateMalformedActualLRP(guid string, index int32) {
	t.createMalformedValueForKey(etcddb.ActualLRPSchemaPath(guid, index))
}

func (t *ETCDHelper) CreateMalformedEvacuatingLRP(guid string, index int32) {
	t.createMalformedValueForKey(etcddb.EvacuatingActualLRPSchemaPath(guid, index))
}

func (t *ETCDHelper) CreateMalformedDesiredLRP(guid string) {
	t.createMalformedValueForKey(etcddb.DesiredLRPSchemaPath(&models.DesiredLRP{ProcessGuid: guid}))
}

func (t *ETCDHelper) CreateMalformedTask(guid string) {
	t.createMalformedValueForKey(etcddb.TaskSchemaPath(&models.Task{TaskGuid: guid}))
}

func (t *ETCDHelper) createMalformedValueForKey(key string) {
	_, err := t.client.Create(key, []byte("ßßßßßß"), 0)

	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("error occurred at key '%s'", key))
}

func (t *ETCDHelper) CreateDesiredLRPsInDomains(domainCounts map[string]int) map[string][]*models.DesiredLRP {
	createdDesiredLRPs := map[string][]*models.DesiredLRP{}

	for domain, count := range domainCounts {
		createdDesiredLRPs[domain] = []*models.DesiredLRP{}

		for i := 0; i < count; i++ {
			action := &models.Action{}
			action.SetValue(&models.DownloadAction{
				From: "http://example.com",
				To:   "/tmp/internet",
				User: "someone",
			})
			desiredLRP := &models.DesiredLRP{
				Domain:      domain,
				ProcessGuid: fmt.Sprintf("guid-%d-for-%s", i, domain),
				RootFs:      "some:rootfs",
				Instances:   1,
				Action:      action,
			}
			value, err := t.serializer.Marshal(t.logger, t.format, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			t.client.Set(etcddb.DesiredLRPSchemaPath(desiredLRP), value, 0)
			Expect(err).NotTo(HaveOccurred())

			createdDesiredLRPs[domain] = append(createdDesiredLRPs[domain], desiredLRP)
		}
	}

	return createdDesiredLRPs
}
