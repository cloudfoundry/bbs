package fakesqldriver_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb"
	"code.cloudfoundry.org/bbs/db/sqldb/fakesqldriver/fakesqldriverfakes"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/bbs/db/sqldb/helpers/monitor"
	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/bbs/format"
	"code.cloudfoundry.org/bbs/guidprovider/guidproviderfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	mfakes "code.cloudfoundry.org/diego-logging-client/testhelpers"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFakesqldriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakesqldriver Suite")
}

var (
	fakeTx           *fakesqldriverfakes.FakeTx
	fakeConn         *fakesqldriverfakes.FakeConn
	fakeDriver       *fakesqldriverfakes.FakeDriver
	fakeClock        *fakeclock.FakeClock
	fakeGUIDProvider *guidproviderfakes.FakeGUIDProvider
	fakeMetronClient *mfakes.FakeIngressClient
	logger           *lagertest.TestLogger

	db         *sql.DB
	cryptor    encryption.Cryptor
	serializer format.Serializer

	sqlDB *sqldb.SQLDB

	ctx context.Context
)

var _ = BeforeEach(func() {
	var err error
	fakeClock = fakeclock.NewFakeClock(time.Now())
	fakeGUIDProvider = &guidproviderfakes.FakeGUIDProvider{}
	fakeMetronClient = new(mfakes.FakeIngressClient)
	logger = lagertest.NewTestLogger("sql-db")

	fakeDriver = &fakesqldriverfakes.FakeDriver{}
	fakeConn = &fakesqldriverfakes.FakeConn{}
	fakeTx = &fakesqldriverfakes.FakeTx{}

	fakeDriver.OpenReturns(fakeConn, nil)

	fakeConn.BeginReturns(fakeTx, nil)

	guid, err := uuid.NewV4()
	Expect(err).NotTo(HaveOccurred())
	driverName := fmt.Sprintf("fake-%s", guid)

	sql.Register(driverName, fakeDriver)

	db, err = sql.Open(driverName, "")
	Expect(err).NotTo(HaveOccurred())
	db.SetMaxIdleConns(0)
	Expect(db.Ping()).NotTo(HaveOccurred())

	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor = encryption.NewCryptor(keyManager, rand.Reader)
	serializer = format.NewSerializer(cryptor)

	helperDB := helpers.NewMonitoredDB(db, monitor.New())
	sqlDB = sqldb.NewSQLDB(helperDB, 5, 5, cryptor, fakeGUIDProvider, fakeClock, helpers.MySQL, fakeMetronClient)

	ctx = context.Background()
})

var _ = AfterEach(func() {
	Expect(db.Close()).To(Succeed())
})
