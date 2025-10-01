package test_helpers

import (
	"fmt"

	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/fixtures"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
)

const (
	MetronCAFile         = "fixtures/metron/CA.crt"
	MetronServerCertFile = "fixtures/metron/metron.crt"
	MetronServerKeyFile  = "fixtures/metron/metron.key"
)

type MetronIngressSetup struct {
	Server            *testhelpers.TestIngressServer
	TestMetricsChan   chan *loggregator_v2.Envelope
	SignalMetricsChan chan struct{}
	Port              int
}

func StartMetronIngress(fixturesPath string) (*MetronIngressSetup, error) {
	fmt.Println("getting metron ...")
	fmt.Println("*** path", fixtures.Path("CA.crt"))
	testIngressServer, err := testhelpers.NewTestIngressServer(fixtures.Path("metron.crt"), fixtures.Path("metron.key"), fixtures.Path("CA.crt"))
	if err != nil {
		return nil, err
	}
	if err := testIngressServer.Start(); err != nil {
		return nil, err
	}

	receiversChan := testIngressServer.Receivers()

	testMetricsChan, signalMetricsChan := testhelpers.TestMetricChan(receiversChan)
	port, err := testIngressServer.Port()
	if err != nil {
		return nil, err
	}

	return &MetronIngressSetup{
		Server:            testIngressServer,
		TestMetricsChan:   testMetricsChan,
		SignalMetricsChan: signalMetricsChan,
		Port:              port,
	}, nil
}

func GetLoggregatorConfigWithMetronCerts() loggingclient.Config {
	return loggingclient.Config{
		CACertPath: fixtures.Path("CA.crt"),
		CertPath:   fixtures.Path("metron.crt"),
		KeyPath:    fixtures.Path("metron.key"),
	}
}
