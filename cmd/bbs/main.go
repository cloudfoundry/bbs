package main

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/auctioneer"
	"github.com/cloudfoundry-incubator/bbs"
	etcddb "github.com/cloudfoundry-incubator/bbs/db/etcd"
	"github.com/cloudfoundry-incubator/bbs/db/migrations"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/encryptor"
	"github.com/cloudfoundry-incubator/bbs/events"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/handlers"
	"github.com/cloudfoundry-incubator/bbs/metrics"
	"github.com/cloudfoundry-incubator/bbs/migration"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/bbs/taskworkpool"
	"github.com/cloudfoundry-incubator/bbs/watcher"
	"github.com/cloudfoundry-incubator/cf-debug-server"
	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry-incubator/consuladapter"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/rep"
	"github.com/cloudfoundry/dropsonde"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/hashicorp/consul/api"
	"github.com/nu7hatch/gouuid"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var listenAddress = flag.String(
	"listenAddress",
	"",
	"The host:port that the server is bound to.",
)

var requireSSL = flag.Bool(
	"requireSSL",
	false,
	"whether the bbs server should require ssl-secured communication",
)

var caFile = flag.String(
	"caFile",
	"",
	"the certificate authority public key file to use with ssl authentication",
)

var certFile = flag.String(
	"certFile",
	"",
	"the public key file to use with ssl authentication",
)

var keyFile = flag.String(
	"keyFile",
	"",
	"the private key file to use with ssl authentication",
)

var advertiseURL = flag.String(
	"advertiseURL",
	"",
	"The URL to advertise to clients",
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
	"bbs",
	"consul session name",
)

var consulCluster = flag.String(
	"consulCluster",
	"",
	"comma-separated list of consul server URLs (scheme://ip:port)",
)

var lockTTL = flag.Duration(
	"lockTTL",
	locket.LockTTL,
	"TTL for service lock",
)

var lockRetryInterval = flag.Duration(
	"lockRetryInterval",
	locket.RetryInterval,
	"interval to wait before retrying a failed lock acquisition",
)

var reportInterval = flag.Duration(
	"metricsReportInterval",
	time.Minute,
	"interval on which to report metrics",
)

var dropsondePort = flag.Int(
	"dropsondePort",
	3457,
	"port the local metron agent is listening on",
)

var convergenceWorkers = flag.Int(
	"convergenceWorkers",
	20,
	"Max concurrency for convergence",
)

var updateWorkers = flag.Int(
	"updateWorkers",
	1000,
	"Max concurrency for etcd updates in a single request",
)

var taskCallBackWorkers = flag.Int(
	"taskCallBackWorkers",
	1000,
	"Max concurrency for task callback requests",
)

var desiredLRPCreationTimeout = flag.Duration(
	"desiredLRPCreationTimeout",
	1*time.Minute,
	"Expected maximum time to create all components of a desired LRP",
)

const (
	dropsondeOrigin = "bbs"

	bbsWatchRetryWaitDuration = 3 * time.Second
)

func main() {
	cf_debug_server.AddFlags(flag.CommandLine)
	cf_lager.AddFlags(flag.CommandLine)
	etcdFlags := AddETCDFlags(flag.CommandLine)
	encryptionFlags := encryption.AddEncryptionFlags(flag.CommandLine)

	flag.Parse()

	cf_http.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cf_lager.New("bbs")
	logger.Info("starting")

	initializeDropsonde(logger)

	clock := clock.NewClock()

	consulClient, err := consuladapter.NewClient(*consulCluster)
	if err != nil {
		logger.Fatal("new-consul-client-failed", err)
	}

	sessionManager := consuladapter.NewSessionManager(consulClient)

	serviceClient := initializeServiceClient(logger, clock, consulClient, sessionManager)

	maintainer := initializeLockMaintainer(logger, serviceClient)

	cbWorkPool := taskworkpool.New(logger, *taskCallBackWorkers, taskworkpool.HandleCompletedTask)

	etcdOptions, err := etcdFlags.Validate()
	if err != nil {
		logger.Fatal("etcd-validation-failed", err)
	}
	storeClient := initializeEtcdStoreClient(logger, etcdOptions)

	key, keys, err := encryptionFlags.Parse()
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	keyManager, err := encryption.NewKeyManager(key, keys)
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)

	db := initializeEtcdDB(logger, cryptor, storeClient, cbWorkPool, serviceClient, *desiredLRPCreationTimeout)

	encryptor := encryptor.New(logger, db, keyManager, cryptor, storeClient, clock)

	migrationsDone := make(chan struct{})

	migrationManager := migration.NewManager(logger,
		db,
		cryptor,
		storeClient,
		migrations.Migrations,
		migrationsDone,
		clock,
	)

	desiredHub := events.NewHub()
	desiredStreamer := watcher.NewDesiredStreamer(db)
	desiredWatcher := watcher.NewWatcher(
		logger,
		"desired-lrps",
		bbsWatchRetryWaitDuration,
		desiredStreamer,
		desiredHub,
		clock,
	)

	actualHub := events.NewHub()
	actualStreamer := watcher.NewActualStreamer(db)
	actualWatcher := watcher.NewWatcher(
		logger,
		"actual-lrps",
		bbsWatchRetryWaitDuration,
		actualStreamer,
		actualHub,
		clock,
	)

	taskHub := events.NewHub()
	taskStreamer := watcher.NewTaskStreamer(db)
	taskWatcher := watcher.NewWatcher(
		logger,
		"tasks",
		bbsWatchRetryWaitDuration,
		taskStreamer,
		taskHub,
		clock,
	)

	handler := handlers.New(logger, db, desiredHub, actualHub, taskHub, serviceClient, migrationsDone)

	metricsNotifier := metrics.NewPeriodicMetronNotifier(
		logger,
		*reportInterval,
		etcdOptions,
		clock,
	)

	var server ifrit.Runner
	if *requireSSL {
		tlsConfig, err := cf_http.NewTLSConfig(*certFile, *keyFile, *caFile)
		if err != nil {
			logger.Fatal("tls-configuration-failed", err)
		}
		server = http_server.NewTLSServer(*listenAddress, handler, tlsConfig)
	} else {
		server = http_server.New(*listenAddress, handler)
	}

	members := grouper.Members{
		{"lock-maintainer", maintainer},
		{"workPool", cbWorkPool},
		{"server", server},
		{"migration-manager", migrationManager},
		{"encryptor", encryptor},
		{"desired-watcher", desiredWatcher},
		{"actual-watcher", actualWatcher},
		{"task-watcher", taskWatcher},
		{"metrics", *metricsNotifier},
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

func initializeLockMaintainer(logger lager.Logger, serviceClient bbs.ServiceClient) ifrit.Runner {
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate uuid", err)
	}

	if *advertiseURL == "" {
		logger.Fatal("Advertise URL must be specified", nil)
	}

	bbsPresence := models.NewBBSPresence(uuid.String(), *advertiseURL)
	lockMaintainer, err := serviceClient.NewBBSLockRunner(logger, &bbsPresence, *lockRetryInterval)
	if err != nil {
		logger.Fatal("Couldn't create lock maintainer", err)
	}

	return lockMaintainer
}

func initializeAuctioneerClient(logger lager.Logger) auctioneer.Client {
	if *auctioneerAddress == "" {
		logger.Fatal("auctioneer-address-validation-failed", errors.New("auctioneerAddress is required"))
	}
	return auctioneer.NewClient(*auctioneerAddress)
}

func initializeDropsonde(logger lager.Logger) {
	dropsondeDestination := fmt.Sprint("localhost:", *dropsondePort)
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed-to-initialize-dropsonde", err)
	}
}

func initializeEtcdDB(
	logger lager.Logger,
	cryptor encryption.Cryptor,
	storeClient etcddb.StoreClient,
	cbClient taskworkpool.TaskCompletionClient,
	serviceClient bbs.ServiceClient,
	desiredLRPCreationMaxTime time.Duration,
) *etcddb.ETCDDB {
	return etcddb.NewETCD(
		format.ENCRYPTED_PROTO,
		*convergenceWorkers,
		*updateWorkers,
		desiredLRPCreationMaxTime,
		cryptor,
		storeClient,
		initializeAuctioneerClient(logger),
		serviceClient,
		clock.NewClock(),
		rep.NewClientFactory(cf_http.NewClient(), cf_http.NewClient()),
		cbClient,
	)
}

func initializeEtcdStoreClient(logger lager.Logger, etcdOptions *etcddb.ETCDOptions) etcddb.StoreClient {
	var etcdClient *etcdclient.Client
	var tr *http.Transport

	if etcdOptions.IsSSL {
		if etcdOptions.CertFile == "" || etcdOptions.KeyFile == "" {
			logger.Fatal("failed-to-construct-etcd-tls-client", errors.New("Require both cert and key path"))
		}

		var err error
		etcdClient, err = etcdclient.NewTLSClient(etcdOptions.ClusterUrls, etcdOptions.CertFile, etcdOptions.KeyFile, etcdOptions.CAFile)
		if err != nil {
			logger.Fatal("failed-to-construct-etcd-tls-client", err)
		}

		tlsCert, err := tls.LoadX509KeyPair(etcdOptions.CertFile, etcdOptions.KeyFile)
		if err != nil {
			logger.Fatal("failed-to-construct-etcd-tls-client", err)
		}

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{tlsCert},
			InsecureSkipVerify: true,
			ClientSessionCache: tls.NewLRUClientSessionCache(etcdOptions.ClientSessionCacheSize),
		}
		tr = &http.Transport{
			TLSClientConfig:     tlsConfig,
			Dial:                etcdClient.DefaultDial,
			MaxIdleConnsPerHost: etcdOptions.MaxIdleConnsPerHost,
		}
		etcdClient.SetTransport(tr)
		etcdClient.AddRootCA(etcdOptions.CAFile)
	} else {
		etcdClient = etcdclient.NewClient(etcdOptions.ClusterUrls)
	}
	etcdClient.SetConsistency(etcdclient.STRONG_CONSISTENCY)

	return etcddb.NewStoreClient(etcdClient)
}

func initializeServiceClient(logger lager.Logger, clock clock.Clock, consulClient *api.Client, sessionManager consuladapter.SessionManager) bbs.ServiceClient {
	consulDBSession, err := consuladapter.NewSessionNoChecks("consul-db", *lockTTL, consulClient, sessionManager)
	if err != nil {
		logger.Fatal("consul-session-failed", err)
	}

	return bbs.NewServiceClient(consulDBSession, clock)
}
