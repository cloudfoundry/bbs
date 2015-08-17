package main

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/bbs/auctionhandlers"
	"github.com/cloudfoundry-incubator/bbs/cellhandlers"
	consuldb "github.com/cloudfoundry-incubator/bbs/db/consul"
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/watcher"
	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/workpool"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pivotal-golang/clock"
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

var auctioneerAddress = flag.String(
	"auctioneerAddress",
	"",
	"The address to the auctioneer api server",
)

var sessionName = flag.String(
	"sessionName",
	"rep",
	"consul session name",
)

var consulCluster = flag.String(
	"consulCluster",
	"",
	"comma-separated list of consul server URLs (scheme://ip:port)",
)

var lockTTL = flag.Duration(
	"lockTTL",
	10*time.Second,
	"TTL for service lock",
)

const (
	dropsondeDestination = "localhost:3457"
	dropsondeOrigin      = "bbs"

	bbsWatchRetryWaitDuration = 3 * time.Second
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

	err = validateAuctioneerFlag()
	if err != nil {
		logger.Fatal("auctioneer-address-validation-failed", err)
	}
	auctioneerClient := auctionhandlers.NewClient(*auctioneerAddress)
	consulSession := initializeConsul(logger)
	consulDB := consuldb.NewConsul(consulSession)
	cellClient := cellhandlers.NewClient()
	cbWorkPool, err := workpool.NewWorkPool(etcddb.TASK_CB_WORKERS)
	if err != nil {
		logger.Fatal("callback-workpool-creation-failed", err)
	}
	db := etcddb.NewETCD(etcdClient, auctioneerClient, cellClient, consulDB, clock.NewClock(), cbWorkPool, etcddb.CompleteTaskWork)
	hub := events.NewHub()
	watcher := watcher.NewWatcher(
		logger,
		db,
		hub,
		clock.NewClock(),
		bbsWatchRetryWaitDuration,
	)

	handler := handlers.New(logger, db, hub)

	workPoolRunner := func(signals <-chan os.Signal, ready chan<- struct{}) error {
		close(ready)
		<-signals
		go cbWorkPool.Stop()

		return nil
	}

	members := grouper.Members{
		{"workPool", ifrit.RunFunc(workPoolRunner)},
		{"watcher", watcher},
		{"server", http_server.New(*serverAddress, handler)},
		{"hub-closer", closeHub(logger.Session("hub-closer"), hub)},
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

func validateAuctioneerFlag() error {
	if *auctioneerAddress == "" {
		return errors.New("auctioneerAddress is required")
	}
	return nil
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func closeHub(logger lager.Logger, hub events.Hub) ifrit.Runner {
	return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		logger.Info("starting")
		defer logger.Info("finished")

		close(ready)
		logger.Info("started")

		<-signals
		logger.Info("shutting-down")
		hub.Close()

		return nil
	})
}

func initializeConsul(logger lager.Logger) *consuladapter.Session {
	client, err := consuladapter.NewClient(*consulCluster)
	if err != nil {
		logger.Fatal("new-client-failed", err)
	}

	sessionMgr := consuladapter.NewSessionManager(client)
	consulSession, err := consuladapter.NewSessionNoChecks(*sessionName, *lockTTL, client, sessionMgr)
	if err != nil {
		logger.Fatal("consul-session-failed", err)
	}
	return consulSession
}
