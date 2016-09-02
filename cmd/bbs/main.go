package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/converger"
	"code.cloudfoundry.org/bbs/db"
	etcddb "code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/db/migrations"
	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/encryptor"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/guidprovider"
	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/bbs/migration"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/taskworkpool"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/rep"
	"github.com/cloudfoundry/dropsonde"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/consul/api"
	"github.com/nu7hatch/gouuid"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var accessLogPath = flag.String(
	"accessLogPath",
	"",
	"Location of the access log",
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

var healthAddress = flag.String(
	"healthAddress",
	"",
	"The host:port that the healthcheck server is bound to.",
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

var databaseConnectionString = flag.String(
	"databaseConnectionString",
	"",
	"SQL database connection string",
)

var maxDatabaseConnections = flag.Int(
	"maxDatabaseConnections",
	200,
	"Max numbers of SQL database connections",
)

var databaseDriver = flag.String(
	"databaseDriver",
	"mysql",
	"SQL database driver name",
)

var sqlCACertFile = flag.String(
	"sqlCACertFile",
	"",
	"SQL database client cert, if supplied, require TLS to SQL",
)

var convergeRepeatInterval = flag.Duration(
	"convergeRepeatInterval",
	30*time.Second,
	"the interval between runs of the converger",
)

var kickTaskDuration = flag.Duration(
	"kickTaskDuration",
	30*time.Second,
	"the interval, in seconds, between kicks to tasks",
)

var expireCompletedTaskDuration = flag.Duration(
	"expireCompletedTaskDuration",
	120*time.Second,
	"completed, unresolved tasks are deleted after this duration",
)

var expirePendingTaskDuration = flag.Duration(
	"expirePendingTaskDuration",
	30*time.Minute,
	"unclaimed tasks are marked as failed, after this duration",
)

const (
	dropsondeOrigin           = "bbs"
	bbsWatchRetryWaitDuration = 3 * time.Second
)

func main() {
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)
	etcdFlags := AddETCDFlags(flag.CommandLine)
	encryptionFlags := encryption.AddEncryptionFlags(flag.CommandLine)

	flag.Parse()

	cfhttp.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cflager.New("bbs")
	logger.Info("starting")

	initializeDropsonde(logger)

	clock := clock.NewClock()

	consulClient, err := consuladapter.NewClientFromUrl(*consulCluster)
	if err != nil {
		logger.Fatal("new-consul-client-failed", err)
	}

	serviceClient := bbs.NewServiceClient(consulClient, clock)

	maintainer := initializeLockMaintainer(logger, serviceClient)

	_, portString, err := net.SplitHostPort(*listenAddress)
	if err != nil {
		logger.Fatal("failed-invalid-listen-address", err)
	}
	portNum, err := net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-listen-port", err)
	}

	_, portString, err = net.SplitHostPort(*healthAddress)
	if err != nil {
		logger.Fatal("failed-invalid-health-address", err)
	}
	_, err = net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-health-port", err)
	}

	registrationRunner := initializeRegistrationRunner(logger, consulClient, portNum, clock)

	cbWorkPool := taskworkpool.New(logger, *taskCallBackWorkers, taskworkpool.HandleCompletedTask)

	var activeDB db.DB
	var sqlDB *sqldb.SQLDB
	var sqlConn *sql.DB
	var storeClient etcddb.StoreClient
	var etcdDB *etcddb.ETCDDB

	key, keys, err := encryptionFlags.Parse()
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	keyManager, err := encryption.NewKeyManager(key, keys)
	if err != nil {
		logger.Fatal("cannot-setup-encryption", err)
	}
	cryptor := encryption.NewCryptor(keyManager, rand.Reader)

	etcdOptions, err := etcdFlags.Validate()
	if err != nil {
		logger.Fatal("etcd-validation-failed", err)
	}

	if etcdOptions.IsConfigured {
		storeClient = initializeEtcdStoreClient(logger, etcdOptions)
		etcdDB = initializeEtcdDB(logger, cryptor, storeClient, cbWorkPool, serviceClient, *desiredLRPCreationTimeout)
		activeDB = etcdDB
	}

	// If SQL database info is passed in, use SQL instead of ETCD
	if *databaseDriver != "" && *databaseConnectionString != "" {
		var err error
		connectionString := appendSSLConnectionStringParam(logger, *databaseDriver, *databaseConnectionString, *sqlCACertFile)

		sqlConn, err = sql.Open(*databaseDriver, connectionString)
		if err != nil {
			logger.Fatal("failed-to-open-sql", err)
		}
		defer sqlConn.Close()
		sqlConn.SetMaxOpenConns(*maxDatabaseConnections)
		sqlConn.SetMaxIdleConns(*maxDatabaseConnections)

		err = sqlConn.Ping()
		if err != nil {
			logger.Fatal("sql-failed-to-connect", err)
		}

		sqlDB = sqldb.NewSQLDB(sqlConn, *convergenceWorkers, *updateWorkers, format.ENCRYPTED_PROTO, cryptor, guidprovider.DefaultGuidProvider, clock, *databaseDriver)
		err = sqlDB.CreateConfigurationsTable(logger)
		if err != nil {
			logger.Fatal("sql-failed-create-configurations-table", err)
		}
		activeDB = sqlDB
	}

	if activeDB == nil {
		logger.Fatal("no-database-configured", errors.New("no database configured"))
	}

	encryptor := encryptor.New(logger, activeDB, keyManager, cryptor, clock)

	migrationsDone := make(chan struct{})

	migrationManager := migration.NewManager(
		logger,
		etcdDB,
		storeClient,
		sqlDB,
		sqlConn,
		cryptor,
		migrations.Migrations,
		migrationsDone,
		clock,
		*databaseDriver,
	)

	desiredHub := events.NewHub()
	actualHub := events.NewHub()

	repClientFactory := rep.NewClientFactory(cfhttp.NewClient(), cfhttp.NewClient())
	auctioneerClient := initializeAuctioneerClient(logger)

	exitChan := make(chan struct{})

	var accessLogger lager.Logger
	if *accessLogPath != "" {
		accessLogger = lager.NewLogger("bbs-access")
		file, err := os.OpenFile(*accessLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Error("invalid-access-log-path", err, lager.Data{"access-log-path": *accessLogPath})
			os.Exit(1)
		}
		accessLogger.RegisterSink(lager.NewWriterSink(file, lager.INFO))
	}

	handler := handlers.New(
		logger,
		accessLogger,
		*updateWorkers,
		*convergenceWorkers,
		activeDB,
		desiredHub,
		actualHub,
		cbWorkPool,
		serviceClient,
		auctioneerClient,
		repClientFactory,
		migrationsDone,
		exitChan,
	)

	metricsNotifier := metrics.NewPeriodicMetronNotifier(
		logger,
		*reportInterval,
		etcdOptions,
		clock,
	)

	retirer := controllers.NewActualLRPRetirer(activeDB, actualHub, repClientFactory, serviceClient)
	lrpConvergenceController := controllers.NewLRPConvergenceController(logger, activeDB, actualHub, auctioneerClient, serviceClient, retirer, *convergenceWorkers)
	taskController := controllers.NewTaskController(activeDB, cbWorkPool, auctioneerClient, serviceClient, repClientFactory)

	convergerProcess := converger.New(
		logger,
		clock,
		lrpConvergenceController,
		taskController,
		serviceClient,
		*convergeRepeatInterval,
		*kickTaskDuration,
		*expirePendingTaskDuration,
		*expireCompletedTaskDuration)

	var server ifrit.Runner
	if *requireSSL {
		tlsConfig, err := cfhttp.NewTLSConfig(*certFile, *keyFile, *caFile)
		if err != nil {
			logger.Fatal("tls-configuration-failed", err)
		}
		server = http_server.NewTLSServer(*listenAddress, handler, tlsConfig)
	} else {
		server = http_server.New(*listenAddress, handler)
	}

	healthcheckServer := http_server.New(*healthAddress, http.HandlerFunc(healthCheckHandler))

	members := grouper.Members{
		{"healthcheck", healthcheckServer},
		{"lock-maintainer", maintainer},
		{"workpool", cbWorkPool},
		{"server", server},
		{"migration-manager", migrationManager},
		{"encryptor", encryptor},
		{"hub-maintainer", hubMaintainer(logger, desiredHub, actualHub)},
		{"metrics", *metricsNotifier},
		{"converger", convergerProcess},
		{"registration-runner", registrationRunner},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
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

func appendSSLConnectionStringParam(logger lager.Logger, driverName, databaseConnectionString, sqlCACertFile string) string {
	switch driverName {
	case "mysql":
		if sqlCACertFile != "" {
			certBytes, err := ioutil.ReadFile(sqlCACertFile)
			if err != nil {
				logger.Fatal("failed-to-read-sql-ca-file", err)
			}

			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
				logger.Fatal("failed-to-parse-sql-ca", err)
			}

			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            caCertPool,
			}

			mysql.RegisterTLSConfig("bbs-tls", tlsConfig)
			databaseConnectionString = fmt.Sprintf("%s?tls=bbs-tls", databaseConnectionString)
		}
	case "postgres":
		if sqlCACertFile == "" {
			databaseConnectionString = fmt.Sprintf("%s?sslmode=disable", databaseConnectionString)
		} else {
			databaseConnectionString = fmt.Sprintf("%s?sslmode=verify-ca&sslrootcert=%s", databaseConnectionString, sqlCACertFile)
		}
	}

	return databaseConnectionString
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func hubMaintainer(logger lager.Logger, desiredHub, actualHub events.Hub) ifrit.RunFunc {
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
		return nil
	}
}

func initializeRegistrationRunner(
	logger lager.Logger,
	consulClient consuladapter.Client,
	port int,
	clock clock.Clock) ifrit.Runner {
	registration := &api.AgentServiceRegistration{
		Name: "bbs",
		Port: port,
		Check: &api.AgentServiceCheck{
			TTL: "3s",
		},
	}
	return locket.NewRegistrationRunner(logger, registration, consulClient, locket.RetryInterval, clock)
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
	lockMaintainer, err := serviceClient.NewBBSLockRunner(logger, &bbsPresence, *lockRetryInterval, *lockTTL)
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
		clock.NewClock(),
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
