package models

func (request *EvacuateRunningActualLRPRequest) SetRoutable(routable bool) {
	request.Routable = &routable
}

func (request *EvacuateRunningActualLRPRequest) GetRoutablePtr() *bool {
	return request.Routable
}

func (request *EvacuateRunningActualLRPRequest) RoutableExists() bool {
	ptr := request.GetRoutablePtr()
	return ptr != nil
}
