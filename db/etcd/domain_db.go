package db

import (
	"path"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/coreos/go-etcd/etcd"
)

const DomainSchemaRoot = DataSchemaRoot + "domain"

func (db *ETCDDB) GetAllDomains() (*models.Domains, error) {
	response, err := db.client.Get(DomainSchemaRoot, false, true)
	if err != nil {
		if err.(*etcd.EtcdError).ErrorCode == 100 {
			return &models.Domains{}, nil
		}
		return nil, err
	}

	domains := []string{}
	for _, child := range response.Node.Nodes {
		domains = append(domains, path.Base(child.Key))
	}

	return &models.Domains{Domains: domains}, nil
}

func (db *ETCDDB) UpsertDomain(domain string, ttl int) error {
	_, err := db.client.Set(DomainSchemaPath(domain), "", uint64(ttl))
	return err
}

func DomainSchemaPath(domain string) string {
	return path.Join(DomainSchemaRoot, domain)
}
