package db

import (
	"path"

	"github.com/coreos/go-etcd/etcd"
)

const DataSchemaRoot = "/v1/"
const DomainSchemaRoot = DataSchemaRoot + "domain"

type ETCDDB struct {
	client *etcd.Client
}

func NewETCD(etcdClient *etcd.Client) *ETCDDB {
	return &ETCDDB{etcdClient}
}

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

func DomainSchemaPath(domain string) string {
	return path.Join(DomainSchemaRoot, domain)
}
