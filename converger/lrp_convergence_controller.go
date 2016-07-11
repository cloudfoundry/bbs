package converger

import "code.cloudfoundry.org/lager"

//go:generate counterfeiter -o fake_controllers/fake_lrp_convergence_controller.go . LrpConvergenceController

type LrpConvergenceController interface {
	ConvergeLRPs(logger lager.Logger) error
}
