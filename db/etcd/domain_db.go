package db

import (
	"path"

	"github.com/coreos/go-etcd/etcd"
)

const DomainSchemaRoot = DataSchemaRoot + "domain"

func (db *ETCDDB) GetAllDomains() ([]string, error) {
	response, err := db.client.Get(DomainSchemaRoot, false, true)
	if err != nil {
		if err.(*etcd.EtcdError).ErrorCode == 100 {
			return []string{}, nil
		}
		return nil, err
	}

	domains := []string{}
	for _, child := range response.Node.Nodes {
		domains = append(domains, path.Base(child.Key))
	}

	return domains, nil
}

func (db *ETCDDB) UpsertDomain(domain string, ttl int) error {
	_, err := db.client.Set(DomainSchemaPath(domain), "", uint64(ttl))
	return err
}

func DomainSchemaPath(domain string) string {
	return path.Join(DomainSchemaRoot, domain)
}
