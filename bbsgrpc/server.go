package bbsgrpc

import (
	"crypto/tls"
	"net"
	"os"

	"code.cloudfoundry.org/bbs/models"

	"google.golang.org/grpc"
)

type bbsGRPCServerRunner struct {
	listenAddress string
	tlsConfig     *tls.Config
	handler       models.BBSServer
}

func NewTLSServer(listenAddress string, handler models.BBSServer, tlsConfig *tls.Config) *bbsGRPCServerRunner {
	return &bbsGRPCServerRunner{
		listenAddress: listenAddress,
		tlsConfig:     tlsConfig,
		handler:       handler,
	}
}

func (s *bbsGRPCServerRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var listener net.Listener

	lis, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}

	if s.tlsConfig == nil {
		listener = lis
	} else {
		listener = tls.NewListener(lis, s.tlsConfig)
	}

	grpcServer := grpc.NewServer()
	models.RegisterBBSServer(grpcServer, s.handler)

	errCh := make(chan error)

	go func() {
		errCh <- grpcServer.Serve(listener)
	}()

	select {
	case <-signals:
		return nil
	case err := <-errCh:
		return err
	}
}
