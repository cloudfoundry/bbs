package handlers_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"code.cloudfoundry.org/diego-db-helpers/guidprovider"
	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers"
	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers/monitor"
	"code.cloudfoundry.org/diego-db-helpers/testhelpers"
	"code.cloudfoundry.org/diego-db-helpers/testhelpers/sqlrunner"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/expiration/expirationfakes"
	"code.cloudfoundry.org/locket/handlers"
	"code.cloudfoundry.org/locket/metrics/helpers/helpersfakes"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("LocketHandler", func() {
	var (
		sqlProcess         ifrit.Process
		sqlRunner          sqlrunner.SQLRunner
		lockDB             *db.SQLDB
		sqlConn            *sql.DB
		fakeLockPick       *expirationfakes.FakeLockPick
		logger             *lagertest.TestLogger
		locketHandler      models.LocketServer
		resource           *models.Resource
		exitCh             chan struct{}
		fakeRequestMetrics *helpersfakes.FakeRequestMetrics
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("locket-handler")

		dbName := fmt.Sprintf("diego_%d", GinkgoParallelProcess())
		sqlRunner = testhelpers.NewSQLRunner(dbName)
		sqlProcess = ginkgomon.Invoke(sqlRunner)

		var err error
		dbParams := &helpers.ConnectParams{
			DriverName:                    sqlRunner.DriverName(),
			DatabaseConnectionString:      sqlRunner.ConnectionString(),
			SqlCACertFile:                 "",
			SqlEnableIdentityVerification: false,
		}
		sqlConn, err = helpers.Connect(logger, dbParams)
		Expect(err).NotTo(HaveOccurred())

		dbMonitor := monitor.New()
		monitoredDB := helpers.NewMonitoredDB(sqlConn, dbMonitor)

		lockDB = db.NewSQLDB(
			monitoredDB,
			sqlRunner.DriverName(),
			guidprovider.DefaultGuidProvider,
		)
		err = lockDB.CreateLockTable(context.Background(), logger)
		Expect(err).NotTo(HaveOccurred())

		fakeLockPick = &expirationfakes.FakeLockPick{}
		fakeRequestMetrics = &helpersfakes.FakeRequestMetrics{}

		exitCh = make(chan struct{}, 1)

		resource = &models.Resource{
			Key:      "test",
			Value:    "test-value",
			Owner:    "myself",
			TypeCode: models.LOCK,
		}

		locketHandler = handlers.NewLocketHandler(
			logger,
			lockDB,
			fakeLockPick,
			fakeRequestMetrics,
			exitCh,
			handlers.DefaultDBOperationTimeout,
		)
	})

	AfterEach(func() {
		Expect(sqlConn.Close()).To(Succeed())
		ginkgomon.Kill(sqlProcess)
	})

	Context("when request is cancelled", func() {
		BeforeEach(func() {
			sqlConn.SetMaxOpenConns(1)
			go func() {
				defer GinkgoRecover()
				var sleepQuery string
				if testhelpers.UseMySQL() {
					sleepQuery = "select sleep(60);"
				} else if testhelpers.UsePostgres() {
					sleepQuery = `--pg_sleep_query_context_test
            select pg_sleep(60);`
				} else {
					Fail("unknown db driver")
				}
				sqlConn.Exec(sleepQuery)
			}()
		})

		AfterEach(func() {
			sqlConn.SetMaxOpenConns(0)
			if testhelpers.UsePostgres() {
				// cancel the sleep query in postgres, since it does not allow to drop the database
				_, err := sqlConn.Exec(`SELECT pg_cancel_backend(pid)
				FROM pg_stat_activity
				WHERE state = 'active'
				AND query LIKE '--pg_sleep_query_context_test%'`)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("does not cancel the database request when the grpc context is cancelled", func() {
			ctxWithCancel, cancelFn := context.WithCancel(context.Background())

			finishedRequest := make(chan error, 1)
			go func() {
				defer GinkgoRecover()
				_, err := locketHandler.Lock(ctxWithCancel, &models.LockRequest{
					Resource:     resource,
					TtlInSeconds: 10,
				})
				finishedRequest <- err
			}()

			Eventually(logger).Should(gbytes.Say("started"))

			cancelFn()

			// The DB operation should NOT be cancelled by the gRPC context.
			// With MaxOpenConns=1 and a pg_sleep(60) blocking the connection,
			// the Lock DB call will block waiting for a connection. If the fix
			// is working, the DB context (context.Background) is NOT cancelled,
			// so the call remains blocked — it does NOT return context.Canceled.
			Consistently(finishedRequest, 2*time.Second).ShouldNot(Receive())
		})
	})
})
