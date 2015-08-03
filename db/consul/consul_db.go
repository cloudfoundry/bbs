package consul

import "github.com/cloudfoundry-incubator/consuladapter"

const (
	LockSchemaRoot = "v1/locks"
	CellSchemaRoot = LockSchemaRoot + "/cell"
)

type ConsulDB struct {
	session *consuladapter.Session
}

func NewConsul(session *consuladapter.Session) *ConsulDB {
	return &ConsulDB{session}
}
