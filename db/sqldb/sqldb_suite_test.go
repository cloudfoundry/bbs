package sqldb_test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	thepackagedb "github.com/cloudfoundry-incubator/bbs/db"
	"github.com/cloudfoundry-incubator/bbs/db/sqldb"
	"github.com/cloudfoundry-incubator/bbs/encryption"
	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/cloudfoundry-incubator/bbs/guidprovider/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	_ "github.com/go-sql-driver/mysql"

	"testing"
)

var (
	db               *sql.DB
	sqlDB            thepackagedb.DB
	fakeClock        *fakeclock.FakeClock
	fakeGUIDProvider *fakes.FakeGUIDProvider
	logger           *lagertest.TestLogger
	cryptor          encryption.Cryptor
	serializer       format.Serializer
)

func TestSql(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "SQL DB Suite")
}

var _ = BeforeSuite(func() {
	var err error
	fakeClock = fakeclock.NewFakeClock(time.Now())
	fakeGUIDProvider = &fakes.FakeGUIDProvider{}
	logger = lagertest.NewTestLogger("sql-db")

	// mysql must be set up on localhost as described in the CONTRIBUTING.md doc
	// in diego-release.
	db, err = sql.Open("mysql", "diego:diego_password@/")
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	db, err = sql.Open("mysql", fmt.Sprintf("diego:diego_password@/diego_%d?parseTime=true", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).NotTo(HaveOccurred())

	createTables(db)

	encryptionKey, err := encryption.NewKey("label", "passphrase")
	Expect(err).NotTo(HaveOccurred())
	keyManager, err := encryption.NewKeyManager(encryptionKey, nil)
	Expect(err).NotTo(HaveOccurred())
	cryptor = encryption.NewCryptor(keyManager, rand.Reader)
	serializer = format.NewSerializer(cryptor)

	sqlDB = sqldb.NewSQLDB(db, 5, format.ENCRYPTED_PROTO, cryptor, fakeGUIDProvider, fakeClock)
})

var _ = AfterEach(func() {
	truncateTables(db)
	fakeGUIDProvider.NextGUIDReturns("", nil)
})

var _ = AfterSuite(func() {
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE diego_%d", GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())

	Expect(db.Close()).NotTo(HaveOccurred())
})

func truncateTables(db *sql.DB) {
	for _, query := range truncateTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

func createTables(db *sql.DB) {
	for _, query := range createTablesSQL {
		result, err := db.Exec(query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RowsAffected()).To(BeEquivalentTo(0))
	}
}

var truncateTablesSQL = []string{
	"TRUNCATE TABLE domains;",
	"TRUNCATE TABLE configurations;",
	"TRUNCATE TABLE tasks;",
	"TRUNCATE TABLE desired_lrps;",
	"TRUNCATE TABLE actual_lrps;",
}

var createTablesSQL = []string{
	createDomainSQL,
	createConfigurationsSQL,
	createTasksSQL,
	createDesiredLRPsSQL,
	createActualLRPsSQL,
}

const createDomainSQL = `CREATE TABLE domains(
	domain VARCHAR(255) PRIMARY KEY,
	expire_time BIGINT DEFAULT 0,

	INDEX(expire_time)
);`

const createConfigurationsSQL = `CREATE TABLE configurations(
	id VARCHAR(255) PRIMARY KEY,
	value VARCHAR(255)
);`

const createTasksSQL = `CREATE TABLE tasks(
	guid VARCHAR(255) PRIMARY KEY,
	domain VARCHAR(255) NOT NULL,
	updated_at BIGINT DEFAULT 0,
	created_at BIGINT DEFAULT 0,
	first_completed_at BIGINT DEFAULT 0,
	state INT,
	cell_id VARCHAR(255) NOT NULL DEFAULT "",
	result TEXT,
	failed BOOL DEFAULT false,
	failure_reason VARCHAR(255) NOT NULL DEFAULT "",
	task_definition BLOB NOT NULL,

	INDEX(domain),
	INDEX(state),
	INDEX(cell_id),
	INDEX(updated_at),
	INDEX(created_at),
	INDEX(first_completed_at)
);`

const createDesiredLRPsSQL = `CREATE TABLE desired_lrps(
	process_guid VARCHAR(255) PRIMARY KEY,
	domain VARCHAR(255) NOT NULL,
	log_guid VARCHAR(255) NOT NULL,
	annotation TEXT,
	instances INT NOT NULL,
	memory_mb INT NOT NULL,
	disk_mb INT NOT NULL,
	rootfs VARCHAR(255) NOT NULL,
	routes BLOB NOT NULL,
	volume_placement BLOB NOT NULL,
	modification_tag_epoch VARCHAR(255) NOT NULL,
	modification_tag_index INT,
	run_info BLOB NOT NULL,

	INDEX(domain)
);`

const createActualLRPsSQL = `CREATE TABLE actual_lrps(
	process_guid VARCHAR(255),
	instance_index INT,
	evacuating BOOL DEFAULT false,
	domain VARCHAR(255) NOT NULL,
	state VARCHAR(255) NOT NULL,
	instance_guid VARCHAR(255) NOT NULL DEFAULT "",
	cell_id VARCHAR(255) NOT NULL DEFAULT "",
	placement_error VARCHAR(255) NOT NULL DEFAULT "",
	since BIGINT DEFAULT 0,
	net_info BLOB NOT NULL,
	modification_tag_epoch VARCHAR(255) NOT NULL,
	modification_tag_index INT,
	crash_count INT NOT NULL DEFAULT 0,
	crash_reason VARCHAR(255) NOT NULL DEFAULT "",
	expire_time BIGINT DEFAULT 0,

	PRIMARY KEY(process_guid, instance_index, evacuating),
	INDEX(domain),
	INDEX(cell_id),
	INDEX(since),
	INDEX(state),
	INDEX(expire_time)
);`

func randStr(strSize int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
