package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs/cmd/bbs/config"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/converger"
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers/monitor"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/encryptor"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/guidprovider"
	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/serviceclient"
	"code.cloudfoundry.org/bbs/taskworkpool"
	cfhttp "code.cloudfoundry.org/cfhttp/v2"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/debugserver"
	loggingclient "code.cloudfoundry.org/diego-logging-client"
	"code.cloudfoundry.org/go-loggregator/v8/runtimeemitter"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerflags"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/jointlock"
	"code.cloudfoundry.org/locket/lock"
	"code.cloudfoundry.org/locket/lockheldmetrics"
	locketmodels "code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/rep"
	"code.cloudfoundry.org/tlsconfig"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var configFilePath = flag.String(
	"config",
	"",
	"The path to the JSON configuration file.",
)

const (
	bbsLockKey = "bbs"
)

func main() {
	flag.Parse()

	bbsConfig, err := config.NewBBSConfig(*configFilePath)
	if err != nil {
		panic(err.Error())
	}

	logger, reconfigurableSink := lagerflags.NewFromConfig(bbsConfig.SessionName, bbsConfig.LagerConfig)
	logger.Info("starting")

	metronClient, err := initializeMetron(logger, bbsConfig)
	if err != nil {
		logger.Error("failed-to-initialize-metron-client", err)
		os.Exit(1)
	}

	clock := clock.NewClock()

	_, portString, err := net.SplitHostPort(bbsConfig.HealthAddress)
	if err != nil {
		logger.Fatal("failed-invalid-health-address", err)
	}
	_, err = net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-health-port", err)
	}

	key, keys, err := bbsConfig.EncryptionConfig.Parse()
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	keyManager, err := encryption.NewKeyManager(key, keys)
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)

	if bbsConfig.DatabaseDriver == "" || bbsConfig.DatabaseConnectionString == "" {
		logger.Fatal("no-database-configured", errors.New("no database configured"))
	}

	sqlConn, err := helpers.Connect(
		logger,
		bbsConfig.DatabaseDriver,
		bbsConfig.DatabaseConnectionString,
		bbsConfig.SQLCACertFile,
		bbsConfig.SQLEnableIdentityVerification,
	)
	if err != nil {
		logger.Fatal("failed-to-open-sql", err)
	}
	defer sqlConn.Close()

	sqlConn.SetMaxOpenConns(bbsConfig.MaxOpenDatabaseConnections)
	sqlConn.SetMaxIdleConns(bbsConfig.MaxIdleDatabaseConnections)

	err = sqlConn.Ping()
	if err != nil {
		logger.Fatal("sql-failed-to-connect", err)
	}

	queryMonitor := monitor.New()
	monitoredDB := helpers.NewMonitoredDB(sqlConn, queryMonitor)
	sqlDB := sqldb.NewSQLDB(
		monitoredDB,
		bbsConfig.ConvergenceWorkers,
		bbsConfig.UpdateWorkers,
		cryptor,
		guidprovider.DefaultGuidProvider,
		clock,
		bbsConfig.DatabaseDriver,
		metronClient,
	)
	err = sqlDB.CreateConfigurationsTable(context.Background(), logger)
	if err != nil {
		logger.Fatal("sql-failed-create-configurations-table", err)
	}

	encryptor := encryptor.New(logger, sqlDB, keyManager, cryptor, clock, metronClient)

	migrationsDone := make(chan struct{})

	migrationManager := migration.NewManager(
		logger,
		sqlDB,
		sqlConn,
		cryptor,
		migrations.AllMigrations(),
		migrationsDone,
		clock,
		bbsConfig.DatabaseDriver,
		metronClient,
	)

	desiredHub := events.NewHub(logger)
	actualHub := events.NewHub(logger)
	actualLRPInstanceHub := events.NewHub(logger)
	taskHub := events.NewHub(logger)

	repTLSConfig := &rep.TLSConfig{
		RequireTLS:      bbsConfig.RepRequireTLS,
		CaCertFile:      bbsConfig.RepCACert,
		CertFile:        bbsConfig.RepClientCert,
		KeyFile:         bbsConfig.RepClientKey,
		ClientCacheSize: bbsConfig.RepClientSessionCacheSize,
	}

	httpClient := cfhttp.NewClient(
		cfhttp.WithRequestTimeout(time.Duration(bbsConfig.CommunicationTimeout)),
	)
	repClientFactory, err := rep.NewClientFactory(httpClient, httpClient, repTLSConfig)
	if err != nil {
		logger.Fatal("new-rep-client-factory-failed", err)
	}

	auctioneerClient := initializeAuctioneerClient(logger, &bbsConfig)

	exitChan := make(chan struct{})

	var accessLogger lager.Logger
	if bbsConfig.AccessLogPath != "" {
		accessLogger = lager.NewLogger("bbs-access")
		file, err := os.OpenFile(bbsConfig.AccessLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Error("invalid-access-log-path", err, lager.Data{"access-log-path": bbsConfig.AccessLogPath})
			os.Exit(1)
		}
		accessLogger.RegisterSink(lager.NewWriterSink(file, lager.INFO))
	}

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(bbsConfig.CertFile, bbsConfig.KeyFile),
	).Server(tlsconfig.WithClientAuthenticationFromFile(bbsConfig.CaFile))
	if err != nil {
		logger.Fatal("tls-configuration-failed", err)
	}
	// the BBS server performs requests as a client
	tlsConfig.RootCAs = tlsConfig.ClientCAs

	cbWorkPool := taskworkpool.New(logger,
		bbsConfig.TaskCallbackWorkers,
		taskworkpool.HandleCompletedTask,
		tlsConfig,
		time.Duration(bbsConfig.CommunicationTimeout))

	locks := []grouper.Member{}

	var locketClient locketmodels.LocketClient
	locketClient, err = locket.NewClient(logger, bbsConfig.ClientLocketConfig)
	if err != nil {
		logger.Fatal("failed-to-create-locket-client", err)
	}

	if bbsConfig.UUID == "" {
		logger.Fatal("invalid-uuid", errors.New("invalid-uuid-from-config"))
	}

	lockIdentifier := &locketmodels.Resource{
		Key:      bbsLockKey,
		Owner:    bbsConfig.UUID,
		TypeCode: locketmodels.LOCK,
		Type:     locketmodels.LockType,
	}

	locks = append(locks, grouper.Member{"sql-lock", lock.NewLockRunner(
		logger,
		locketClient,
		lockIdentifier,
		locket.DefaultSessionTTLInSeconds,
		clock,
		locket.SQLRetryInterval,
	)})

	var lock ifrit.Runner
	switch len(locks) {
	case 0:
		logger.Fatal("no-locks-configured", errors.New("Lock configuration must be provided"))
	case 1:
		lock = locks[0]
	default:
		lock = jointlock.NewJointLock(clock, 2*time.Second, locks...)
	}

	serviceClient := serviceclient.NewServiceClient(locketClient)

	logger.Info("report-interval", lager.Data{"value": bbsConfig.ReportInterval})
	fileDescriptorTicker := clock.NewTicker(time.Duration(bbsConfig.ReportInterval))
	requestStatsTicker := clock.NewTicker(time.Duration(bbsConfig.ReportInterval))
	locksHeldTicker := clock.NewTicker(time.Duration(bbsConfig.ReportInterval))

	fileDescriptorPath := fmt.Sprintf("/proc/%d/fd", os.Getpid())
	fileDescriptorMetronNotifier := metrics.NewFileDescriptorMetronNotifier(logger, fileDescriptorTicker, metronClient, fileDescriptorPath)
	requestStatMetronNotifier := metrics.NewRequestStatMetronNotifier(logger, requestStatsTicker, metronClient)
	lockHeldMetronNotifier := lockheldmetrics.NewLockHeldMetronNotifier(logger, locksHeldTicker, metronClient)
	taskStatMetronNotifier := metrics.NewTaskStatMetronNotifier(logger, clock, metronClient)
	dbStatMetronNotifier := metrics.NewDBStatMetronNotifier(logger, clock, monitoredDB, metronClient, queryMonitor)

	handler := handlers.New(
		logger,
		accessLogger,
		bbsConfig.UpdateWorkers,
		bbsConfig.ConvergenceWorkers,
		bbsConfig.MaxTaskRetries,
		requestStatMetronNotifier,
		sqlDB,
		desiredHub,
		actualHub,
		actualLRPInstanceHub,
		taskHub,
		cbWorkPool,
		serviceClient,
		auctioneerClient,
		repClientFactory,
		taskStatMetronNotifier,
		migrationsDone,
		exitChan,
	)

	bbsElectionMetronNotifier := metrics.NewBBSElectionMetronNotifier(logger, metronClient)

	actualLRPController := controllers.NewActualLRPLifecycleController(
		sqlDB,
		sqlDB,
		sqlDB,
		sqlDB,
		auctioneerClient,
		serviceClient,
		repClientFactory,
		actualHub,
		actualLRPInstanceHub,
	)

	lrpStatMetronNotifier := metrics.NewLRPStatMetronNotifier(logger, clock, metronClient)

	lrpConvergenceController := controllers.NewLRPConvergenceController(
		logger,
		clock,
		sqlDB,
		sqlDB,
		sqlDB,
		actualHub,
		actualLRPInstanceHub,
		auctioneerClient,
		serviceClient,
		repClientFactory,
		actualLRPController,
		bbsConfig.ConvergenceWorkers,
		lrpStatMetronNotifier,
	)

	taskController := controllers.NewTaskController(
		sqlDB,
		cbWorkPool,
		auctioneerClient,
		serviceClient,
		repClientFactory,
		taskHub,
		taskStatMetronNotifier,
		bbsConfig.MaxTaskRetries,
	)

	convergerProcess := converger.New(
		logger,
		clock,
		lrpConvergenceController,
		taskController,
		serviceClient,
		time.Duration(bbsConfig.ConvergeRepeatInterval),
		time.Duration(bbsConfig.KickTaskDuration),
		time.Duration(bbsConfig.ExpirePendingTaskDuration),
		time.Duration(bbsConfig.ExpireCompletedTaskDuration),
	)

	var server ifrit.Runner
	if tlsConfig != nil {
		server = http_server.NewTLSServer(bbsConfig.ListenAddress, handler, tlsConfig)
	} else {
		server = http_server.New(bbsConfig.ListenAddress, handler)
	}

	healthcheckServer := http_server.New(bbsConfig.HealthAddress, http.HandlerFunc(healthCheckHandler))

	members := grouper.Members{
		{"healthcheck", healthcheckServer},
		{"periodic-filedescriptor-metrics", fileDescriptorMetronNotifier},
		{"lock-held-metrics", lockHeldMetronNotifier},
		{"lock", lock},
		{"set-lock-held-metrics", lockheldmetrics.SetLockHeldRunner(logger, *lockHeldMetronNotifier)},
		{"workpool", cbWorkPool},
		{"server", server},
		{"migration-manager", migrationManager},
		{"encryptor", encryptor},
		{"hub-maintainer", hubMaintainer(logger, desiredHub, actualHub, taskHub)},
		{"bbs-election-metrics", bbsElectionMetronNotifier},
		{"periodic-metrics", requestStatMetronNotifier},
		{"converger", convergerProcess},
		{"lrp-stat-metron-notifier", lrpStatMetronNotifier},
		{"task-stat-metron-notifier", taskStatMetronNotifier},
		{"db-stat-metron-notifier", dbStatMetronNotifier},
	}

	if bbsConfig.DebugAddress != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(bbsConfig.DebugAddress, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))
	go func() {
		// If a handler writes to this channel, we've hit an unrecoverable error
		// and should shut down (cleanly)
		<-exitChan
		monitor.Signal(os.Interrupt)
	}()

	logger.Info("started")

	err = <-monitor.Wait()
	if sqlConn != nil {
		sqlConn.Close()
	}
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func hubMaintainer(logger lager.Logger, desiredHub, actualHub, taskHub events.Hub) ifrit.RunFunc {
	return func(signals <-chan os.Signal, ready chan<- struct{}) error {
		logger := logger.Session("hub-maintainer")
		close(ready)
		logger.Info("started")
		defer logger.Info("finished")

		<-signals
		err := desiredHub.Close()
		if err != nil {
			logger.Error("error-closing-desired-hub", err)
		}
		err = actualHub.Close()
		if err != nil {
			logger.Error("error-closing-actual-hub", err)
		}
		err = taskHub.Close()
		if err != nil {
			logger.Error("error-closing-actual-hub", err)
		}
		return nil
	}
}

func initializeAuctioneerClient(logger lager.Logger, bbsConfig *config.BBSConfig) auctioneer.Client {
	if bbsConfig.AuctioneerAddress == "" {
		logger.Fatal("auctioneer-address-validation-failed", errors.New("auctioneerAddress is required"))
	}

	if bbsConfig.AuctioneerCACert != "" || bbsConfig.AuctioneerClientCert != "" || bbsConfig.AuctioneerClientKey != "" {
		client, err := auctioneer.NewSecureClient(bbsConfig.AuctioneerAddress,
			bbsConfig.AuctioneerCACert,
			bbsConfig.AuctioneerClientCert,
			bbsConfig.AuctioneerClientKey,
			bbsConfig.AuctioneerRequireTLS,
			time.Duration(bbsConfig.CommunicationTimeout),
		)
		if err != nil {
			logger.Fatal("failed-to-construct-auctioneer-client", err)
		}
		return client
	}

	return auctioneer.NewClient(bbsConfig.AuctioneerAddress, time.Duration(bbsConfig.CommunicationTimeout))
}

func initializeMetron(logger lager.Logger, bbsConfig config.BBSConfig) (loggingclient.IngressClient, error) {
	client, err := loggingclient.NewIngressClient(bbsConfig.LoggregatorConfig)
	if err != nil {
		return nil, err
	}

	if bbsConfig.LoggregatorConfig.UseV2API {
		emitter := runtimeemitter.NewV1(client)
		go emitter.Run()
	}

	return client, nil
}
