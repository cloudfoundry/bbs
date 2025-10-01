package test_helpers

import (
	"path"

	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
)

type MetronIngressSetup struct {
	Server            *testhelpers.TestIngressServer
	TestMetricsChan   chan *loggregator_v2.Envelope
	SignalMetricsChan chan struct{}
	Port              int
}

func StartMetronIngress(fixturesPath string) (*MetronIngressSetup, error) {
	metronCAFile := path.Join(fixturesPath, "metron", "CA.crt")
	metronServerCertFile := path.Join(fixturesPath, "metron", "metron.crt")
	metronServerKeyFile := path.Join(fixturesPath, "metron", "metron.key")

	testIngressServer, err := testhelpers.NewTestIngressServer(metronServerCertFile, metronServerKeyFile, metronCAFile)
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
		CACertPath: "fixtures/metron/CA.crt",
		CertPath:   "fixtures/metron/metron.crt",
		KeyPath:    "fixtures/metron/metron.key",
	}
}
