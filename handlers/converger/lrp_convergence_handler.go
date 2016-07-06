package converger

//go:generate counterfeiter -o fake_handlers/fake_lrp_convergence_handler.go . LrpConvergenceHandler

type LrpConvergenceHandler interface {
	ConvergeLRPs() error
}
