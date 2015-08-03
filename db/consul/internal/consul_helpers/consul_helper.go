package consul_helpers

import "github.com/cloudfoundry-incubator/consuladapter"

type ConsulHelper struct {
	consulSession *consuladapter.Session
}

func NewConsulHelper(consulSession *consuladapter.Session) *ConsulHelper {
	return &ConsulHelper{consulSession: consulSession}
}
