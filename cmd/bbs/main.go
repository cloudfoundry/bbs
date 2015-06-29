package main

import (
	"flag"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry/dropsonde"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var serverAddress = flag.String(
	"address",
	"",
	"The host:port that the server is bound to.",
)

var communicationTimeout = flag.Duration(
	"communicationTimeout",
	10*time.Second,
	"Timeout applied to all HTTP requests.",
)

const (
	dropsondeDestination = "localhost:3457"
	dropsondeOrigin      = "bbs"
)

func main() {
	cf_debug_server.AddFlags(flag.CommandLine)
	cf_lager.AddFlags(flag.CommandLine)
	etcdFlags := AddETCDFlags(flag.CommandLine)
	flag.Parse()

	cf_http.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cf_lager.New("bbs")
	logger.Info("starting")

	initializeDropsonde(logger)

	etcdOptions, err := etcdFlags.Validate()
	if err != nil {
		logger.Fatal("etcd-validation-failed", err)
	}

	var etcdClient *etcdclient.Client
	if etcdOptions.IsSSL {
		etcdClient, err = etcdclient.NewTLSClient(etcdOptions.ClusterUrls, etcdOptions.CertFile, etcdOptions.KeyFile, etcdOptions.CAFile)
		if err != nil {
			logger.Fatal("failed-to-construct-etcd-tls-client", err)
		}
	} else {
		etcdClient = etcdclient.NewClient(etcdOptions.ClusterUrls)
	}
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)
	db := db.NewETCD(etcdClient)
	handler := handlers.New(db, logger)

	members := grouper.Members{
		{"server", http_server.New(*serverAddress, handler)},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}
